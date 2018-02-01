// +build ide test_systemd

package agent_test

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/soil/fixture"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"testing"
	"text/template"
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
	fixture.DestroyUnits("pod-*", "unit-*")
	defer fixture.DestroyUnits("pod-*", "unit-*")

	os.RemoveAll("testdata/.test_server.hcl")
	serverOptions := agent.ServerOptions{
		ConfigPath: []string{
			"testdata/.test_server.hcl",
		},
		Meta:    map[string]string{},
		Address: fmt.Sprintf(":%d", fixture.RandomPort(t)),
	}
	server := agent.NewServer(context.Background(), logx.GetLog("test"), serverOptions)
	require.NoError(t, server.Open())

	allUnitNames := []string{
		"pod-*",
		"unit-*",
	}

	t.Run("0 pods should not be present", func(t *testing.T) {
		//t.Skip()
		fixture.WaitNoErrorT10(t, fixture.UnitStatesFn(allUnitNames, map[string]string{}))
	})
	t.Run("1 deploy first configuration", func(t *testing.T) {
		//t.Skip()
		writeConfig(t, "testdata/server_test_1.hcl", nil)
		server.Configure()
		fixture.WaitNoErrorT10(t, fixture.UnitStatesFn(allUnitNames, map[string]string{
			"pod-private-1.service": "active",
			"pod-private-2.service": "active",
			"unit-1.service":        "active",
			"unit-2.service":        "active",
		}))
	})
	t.Run("2 remove 2 from meta", func(t *testing.T) {
		//t.Skip()
		writeConfig(t, "testdata/server_test_2.hcl", nil)
		server.Configure()

		fixture.WaitNoErrorT10(t, fixture.UnitStatesFn(allUnitNames, map[string]string{
			"pod-private-1.service": "active",
			"unit-1.service":        "active",
		}))
	})
	t.Run("3 ping", func(t *testing.T) {
		//t.Skip()
		res, err := http.Get(fmt.Sprintf("http://127.0.0.1%s/v1/status/ping", serverOptions.Address))
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, res.StatusCode, 200)
	})
	t.Run("4 reload", func(t *testing.T) {
		//t.Skip()
		writeConfig(t, "testdata/server_test_4.hcl", nil)
		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://127.0.0.1%s/v1/agent/reload", serverOptions.Address), nil)
		assert.NoError(t, err)
		_, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)

		fixture.WaitNoErrorT10(t, fixture.UnitStatesFn(allUnitNames, map[string]string{
			"pod-private-1.service": "active",
			"pod-private-2.service": "active",
			"unit-1.service":        "active",
			"unit-2.service":        "active",
		}))
	})
	t.Run("5 drain on", func(t *testing.T) {
		//t.Skip()
		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://127.0.0.1%s/v1/agent/drain", serverOptions.Address), nil)
		assert.NoError(t, err)
		_, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)

		fixture.WaitNoErrorT10(t, fixture.UnitStatesFn(allUnitNames, map[string]string{}))
	})
	t.Run("6 drain off", func(t *testing.T) {
		//t.Skip()
		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://127.0.0.1%s/v1/agent/drain", serverOptions.Address), nil)
		assert.NoError(t, err)
		_, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)

		fixture.WaitNoErrorT10(t, fixture.UnitStatesFn(allUnitNames, map[string]string{
			"pod-private-1.service": "active",
			"pod-private-2.service": "active",
			"unit-1.service":        "active",
			"unit-2.service":        "active",
		}))
	})
	t.Run("8 with dependency failed", func(t *testing.T) {
		//t.Skip()
		writeConfig(t, "testdata/server_test_8.hcl", nil)
		server.Configure()
		fixture.WaitNoErrorT10(t, fixture.UnitStatesFn(allUnitNames, map[string]string{}))
	})
	t.Run("9 with dependency ok", func(t *testing.T) {
		//t.Skip()
		writeConfig(t, "testdata/server_test_9.hcl", nil)
		server.Configure()
		fixture.WaitNoErrorT10(t, fixture.UnitStatesFn(allUnitNames, map[string]string{
			"pod-private-1.service": "active",
			"unit-1.service":        "active",
			"pod-private-2.service": "active",
			"unit-2.service":        "active",
		}))
	})
	t.Run("10 with resource", func(t *testing.T) {
		//t.Skip()
		writeConfig(t, "testdata/server_test_10.hcl", nil)
		server.Configure()
		fixture.WaitNoErrorT10(t, fixture.UnitStatesFn(allUnitNames, map[string]string{
			"pod-private-r1.service": "active",
			"unit-0.service":         "active",
			"pod-private-r2.service": "active",
			"unit-2.service":         "active",
		}))
	})

	server.Close()
	server.Wait()
}
