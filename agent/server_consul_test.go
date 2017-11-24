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
	"github.com/akaspin/soil/lib"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/soil/proto"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"reflect"
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

	cli, cliErr := api.NewClient(&api.Config{
		Address: consulServer.Address(),
	})
	assert.NoError(t, cliErr)

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
		require.NoError(t, server.Open())
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
				var node proto.NodeInfo
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
			resp, err := http.Get(fmt.Sprintf("http://%s/v1/status/ping?node=node", configEnv["AgentAddress"]))
			if err != nil {
				return
			}
			if resp == nil {
				err = fmt.Errorf(`no resp`)
				return
			}
			if resp.StatusCode != 200 {
				err = fmt.Errorf(`bad status code: %d`, resp.StatusCode)
			}
			return
		})
	})
	t.Run(`get nodes`, func(t *testing.T) {
		fixture.WaitNoError(t, waitConfig, func() (err error) {
			resp, err := http.Get(fmt.Sprintf("http://%s/v1/status/nodes", configEnv["AgentAddress"]))
			if err != nil {
				return
			}
			if resp == nil {
				err = fmt.Errorf(`no resp`)
				return
			}
			if resp.StatusCode != 200 {
				err = fmt.Errorf(`bad status code: %d`, resp.StatusCode)
				return
			}
			var v proto.NodesInfo
			if err = json.NewDecoder(resp.Body).Decode(&v); err != nil {
				return
			}
			defer resp.Body.Close()
			if len(v) == 0 {
				err = fmt.Errorf(`no nodes`)
				return
			}
			if v[0].ID != "node" {
				err = fmt.Errorf(`bad node id: %v`, v)
			}
			return
		})
	})
	t.Run(`10 put /v1/registry`, func(t *testing.T) {
		var pods manifest.Registry
		rs := &lib.StaticBuffers{}
		require.NoError(t, rs.ReadFiles("testdata/TestServer_Configure_Consul_10.hcl"))
		require.NoError(t, pods.Unmarshal(manifest.PublicNamespace, rs.GetReaders()...))

		buf := &bytes.Buffer{}
		assert.NoError(t, json.NewEncoder(buf).Encode(pods))
		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://%s/v1/registry", configEnv["AgentAddress"]), bytes.NewReader(buf.Bytes()))
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, resp.StatusCode, 200)
	})
	t.Run(`get /v1/registry`, func(t *testing.T) {
		var pods manifest.Registry
		rs := &lib.StaticBuffers{}
		require.NoError(t, rs.ReadFiles("testdata/TestServer_Configure_Consul_10.hcl"))
		require.NoError(t, pods.Unmarshal(manifest.PublicNamespace, rs.GetReaders()...))

		fixture.WaitNoError10(t, func() (err error) {
			resp, err := http.Get(fmt.Sprintf("http://%s/v1/registry", configEnv["AgentAddress"]))
			if err != nil {
				return
			}
			if resp == nil {
				err = fmt.Errorf(`response is nil`)
				return
			}
			if resp.StatusCode != 200 {
				err = fmt.Errorf(`bad status code: %d`, resp.StatusCode)
				return
			}
			var res manifest.Registry
			if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
				return
			}
			defer resp.Body.Close()
			if !reflect.DeepEqual(res, pods) {
				err = fmt.Errorf(`not equal: (expect)%v != (actual)%v`, pods, res)
			}
			return
		})
	})
	t.Run(`ensure public pods`, func(t *testing.T) {
		fixture.WaitNoError(t, waitConfig, sd.UnitStatesFn(allUnitNames, map[string]string{
			"pod-private-1.service":       "active",
			"unit-1.service":              "active",
			"pod-public-1-public.service": "active",
			"unit-1-public.service":       "active",
			"pod-public-2-public.service": "active",
			"unit-2-public.service":       "active",
		}))
	})
	t.Run(`delete /v1/registry?pod=2-public`, func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://%s/v1/registry?pod=2-public", configEnv["AgentAddress"]), nil)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, resp.StatusCode, 200)
	})
	t.Run(`11 get /v1/registry`, func(t *testing.T) {
		var pods manifest.Registry
		rs := &lib.StaticBuffers{}
		require.NoError(t, rs.ReadFiles("testdata/TestServer_Configure_Consul_11.hcl"))
		require.NoError(t, pods.Unmarshal(manifest.PublicNamespace, rs.GetReaders()...))

		fixture.WaitNoError10(t, func() (err error) {
			resp, err := http.Get(fmt.Sprintf("http://%s/v1/registry", configEnv["AgentAddress"]))
			if err != nil {
				return
			}
			if resp == nil {
				err = fmt.Errorf(`response is nil`)
				return
			}
			if resp.StatusCode != 200 {
				err = fmt.Errorf(`bad status code: %d`, resp.StatusCode)
				return
			}
			var res manifest.Registry
			if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
				return
			}
			defer resp.Body.Close()
			if !reflect.DeepEqual(res, pods) {
				err = fmt.Errorf(`not equal: (expect)%v != (actual)%v`, pods, res)
			}
			return
		})
	})
	t.Run(`ensure 2-public is removed`, func(t *testing.T) {
		fixture.WaitNoError(t, waitConfig, sd.UnitStatesFn(allUnitNames, map[string]string{
			"pod-private-1.service":       "active",
			"unit-1.service":              "active",
			"pod-public-1-public.service": "active",
			"unit-1-public.service":       "active",
		}))
	})
}
