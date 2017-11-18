// +build ide test_systemd,test_cluster

package agent_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/soil/fixture"
	"github.com/akaspin/soil/proto"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"net/http"
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
	consulServer.Up()
	consulServer.WaitAlive()
	consulServer.Pause()

	cli, err := api.NewClient(&api.Config{
		Address: consulServer.Address(),
	})
	assert.NoError(t, err)

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
	t.Run(`unpause consul server`, func(t *testing.T) {
		consulServer.Unpause()
	})
	t.Run(`ensure node announced`, func(t *testing.T) {
		fixture.WaitNoError(t, waitConfig, func() (err error) {
			res, _, err := cli.KV().List("soil/nodes", nil)
			if err != nil {
				return
			}
			if len(res) != 1 {
				err = fmt.Errorf(`node registration not found`)
			}
			var found bool
			for _, raw := range res {
				var node proto.ClusterNode
				if err = json.NewDecoder(bytes.NewReader(raw.Value)).Decode(&node); err != nil {
					return
				}
				if node.ID == "node" && node.Advertise == configEnv["AgentAddress"] {
					found = true
					break
				}
			}
			if !found {
				err = fmt.Errorf(`node not found`)
			}
			return
		})
	})
	t.Run(`ping node`, func(t *testing.T) {
		fixture.WaitNoError(t, waitConfig, func() (err error) {
			_, err = http.Get(fmt.Sprintf("http://%s/status/ping?node=node", configEnv["AgentAddress"]))
			return
		})
	})
}
