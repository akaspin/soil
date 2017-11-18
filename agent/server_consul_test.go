// +build ide test_systemd,test_cluster

package agent_test

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/soil/fixture"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestServer_Configure_Consul(t *testing.T) {
	sd := fixture.NewSystemd("/run/systemd/system", "pod")
	sd.Cleanup()
	defer sd.Cleanup()

	os.RemoveAll("testdata/.test_server.hcl")

	log := logx.GetLog("test")
	serverOptions := agent.ServerOptions{
		ConfigPath: []string{
			"testdata/.test_server.hcl",
		},
		Meta:    map[string]string{},
		Address: fmt.Sprintf(":%d", fixture.RandomPort(t)),
	}
	server := agent.NewServer(context.Background(), log, serverOptions)
	defer server.Close()

	consulServer := fixture.NewConsulServer(t, nil)
	defer consulServer.Clean()

	configEnv := map[string]interface{}{
		"ConsulAddress": consulServer.Address(),
		"AgentAddress":  fmt.Sprintf("%s%s", fixture.GetLocalIP(t), serverOptions.Address),
	}
	waitConfig := fixture.DefaultWaitConfig()
	allUnitNames := []string{
		"pod-*",
		"unit-*",
	}

	t.Run(`start agent`, func(t *testing.T) {
		assert.NoError(t, server.Open())
	})
	t.Run(`0 configure with consul`, func(t *testing.T) {
		writeConfig(t, "testdata/TestServer_Configure_Consul_0.hcl", configEnv)
		server.Configure()
		fixture.WaitNoError(t, waitConfig, sd.UnitStatesFn(allUnitNames, map[string]string{
			"pod-private-1.service": "active",
			"unit-1.service":        "active",
		}))
	})
}
