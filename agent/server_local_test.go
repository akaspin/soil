// +build ide test_systemd

package agent_test

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/soil/fixture"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
	"text/template"
	"time"
)

func writeConfig(t *testing.T, source string, env map[string]interface{}) {
	t.Helper()
	os.RemoveAll("testdata/.test_server.hcl")
	tmpl, err := template.ParseFiles(source)
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}
	f, err := os.Create("testdata/.test_server.hcl")
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}
	defer f.Close()
	if err = tmpl.Execute(f, env); err != nil {
		t.Error(err)
		t.FailNow()
	}
}

func TestServer_Configure_Local(t *testing.T) {
	sd := fixture.NewSystemd("/run/systemd/system", "pod")
	sd.Cleanup()
	defer sd.Cleanup()

	os.RemoveAll("testdata/.test_server.hcl")
	serverOptions := agent.ServerOptions{
		ConfigPath: []string{
			"testdata/.test_server.hcl",
		},
		Meta:    map[string]string{},
		Address: fmt.Sprintf(":%d", fixture.RandomPort(t)),
	}
	server := agent.NewServer(context.Background(), logx.GetLog("test"), serverOptions)
	assert.NoError(t, server.Open())

	allUnitNames := []string{
		"pod-*",
		"unit-*",
	}

	waitConfig := fixture.DefaultWaitConfig()

	t.Run("0 pods should not be present", func(t *testing.T) {
		fixture.WaitNoError(t, waitConfig, sd.UnitStatesFn(allUnitNames, map[string]string{}))
	})
	t.Run("1 deploy first configuration", func(t *testing.T) {
		writeConfig(t, "testdata/server_test_1.hcl", nil)
		server.Configure()
		fixture.WaitNoError(t, waitConfig, sd.UnitStatesFn(allUnitNames, map[string]string{
			"pod-private-1.service": "active",
			"pod-private-2.service": "active",
			"unit-1.service":        "active",
			"unit-2.service":        "active",
		}))
		sd.AssertUnitHashes(t, allUnitNames, map[string]uint64{
			"/run/systemd/system/pod-private-1.service": 0xf114f766af424710,
			"/etc/systemd/system/pod-private-2.service": 0xf8bc5d840f0f6b52,
			"/run/systemd/system/unit-1.service":        0x7f15d00cb10c1836,
			"/etc/systemd/system/unit-2.service":        0xfef5c98efe4f711f,
		})
	})
	t.Run("2 remove 2 from meta", func(t *testing.T) {
		writeConfig(t, "testdata/server_test_2.hcl", nil)
		server.Configure()

		fixture.WaitNoError(t, waitConfig, sd.UnitStatesFn(allUnitNames, map[string]string{
			"pod-private-1.service": "active",
			"unit-1.service":        "active",
		}))
		sd.AssertUnitHashes(t, allUnitNames, map[string]uint64{
			"/run/systemd/system/pod-private-1.service": 0xce80849ad12813cc,
			"/run/systemd/system/unit-1.service":        0xce7b239c1e94def4,
		})
	})
	t.Run("3 ping", func(t *testing.T) {
		res, err := http.Get(fmt.Sprintf("http://127.0.0.1%s/v1/status/ping", serverOptions.Address))
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, res.StatusCode, 200)
	})
	t.Run("4 reload", func(t *testing.T) {
		writeConfig(t, "testdata/server_test_4.hcl", nil)
		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://127.0.0.1%s/v1/agent/reload", serverOptions.Address), nil)
		assert.NoError(t, err)
		_, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)

		fixture.WaitNoError(t, waitConfig, sd.UnitStatesFn(allUnitNames, map[string]string{
			"pod-private-1.service": "active",
			"pod-private-2.service": "active",
			"unit-1.service":        "active",
			"unit-2.service":        "active",
		}))
		sd.AssertUnitHashes(t, allUnitNames, map[string]uint64{
			"/run/systemd/system/pod-private-1.service": 0xf114f766af424710,
			"/etc/systemd/system/pod-private-2.service": 0xf8bc5d840f0f6b52,
			"/run/systemd/system/unit-1.service":        0x7f15d00cb10c1836,
			"/etc/systemd/system/unit-2.service":        0xfef5c98efe4f711f,
		})
	})
	t.Run("5 drain on", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://127.0.0.1%s/v1/agent/drain", serverOptions.Address), nil)
		assert.NoError(t, err)
		_, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)

		fixture.WaitNoError(t, waitConfig, sd.UnitStatesFn(allUnitNames, map[string]string{}))
		sd.AssertUnitHashes(t, allUnitNames, map[string]uint64{})
	})
	t.Run("6 drain off", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://127.0.0.1%s/v1/agent/drain", serverOptions.Address), nil)
		assert.NoError(t, err)
		_, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)

		fixture.WaitNoError(t, waitConfig, sd.UnitStatesFn(allUnitNames, map[string]string{
			"pod-private-1.service": "active",
			"pod-private-2.service": "active",
			"unit-1.service":        "active",
			"unit-2.service":        "active",
		}))
		sd.AssertUnitHashes(t, allUnitNames, map[string]uint64{
			"/run/systemd/system/pod-private-1.service": 0xf114f766af424710,
			"/etc/systemd/system/pod-private-2.service": 0xf8bc5d840f0f6b52,
			"/run/systemd/system/unit-1.service":        0x7f15d00cb10c1836,
			"/etc/systemd/system/unit-2.service":        0xfef5c98efe4f711f,
		})
	})
	t.Run("7 with resource", func(t *testing.T) {
		writeConfig(t, "testdata/server_test_7.hcl", nil)
		server.Configure()
		fixture.WaitNoError(t, fixture.WaitConfig{
			Retry:   time.Millisecond * 500,
			Retries: 100,
		}, sd.UnitStatesFn(allUnitNames, map[string]string{
			"pod-private-1.service": "active",
			"unit-1.service":        "active",
		}))
		sd.AssertUnitHashes(t, allUnitNames, map[string]uint64{
			"/run/systemd/system/pod-private-1.service": 0x9e2aa3b3b95275df,
			"/run/systemd/system/unit-1.service":        0x5ea112942f0c47e8,
		})
	})
	t.Run("8 with dependency failed", func(t *testing.T) {
		writeConfig(t, "testdata/server_test_8.hcl", nil)
		server.Configure()
		fixture.WaitNoError(t, waitConfig, sd.UnitStatesFn(allUnitNames, map[string]string{}))
	})
	t.Run("9 with dependency ok", func(t *testing.T) {
		writeConfig(t, "testdata/server_test_9.hcl", nil)
		server.Configure()
		fixture.WaitNoError(t, waitConfig, sd.UnitStatesFn(allUnitNames, map[string]string{
			"pod-private-1.service": "active",
			"unit-1.service":        "active",
			"pod-private-2.service": "active",
			"unit-2.service":        "active",
		}))
	})

	server.Close()
	server.Wait()

}
