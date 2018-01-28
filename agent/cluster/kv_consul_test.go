// +build ide test_cluster

package cluster_test

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/cluster"
	"github.com/akaspin/soil/fixture"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestKV_ConsulBackend(t *testing.T) {
	// Start KV before Consul
	srv := fixture.NewConsulServer(t, nil)
	defer srv.Clean()
	waitConfig := fixture.DefaultWaitConfig()

	ctx, _ := context.WithCancel(context.Background())
	kv := cluster.NewKV(ctx, logx.GetLog("test"), cluster.DefaultBackendFactory)

	//watcherCtx, _ := context.WithCancel(context.Background())
	//watcher := bus.NewTestingConsumer(ctx)

	cli, err := api.NewClient(&api.Config{
		Address: srv.Address(),
	})
	assert.NoError(t, err)

	t.Run(`start KV`, func(t *testing.T) {
		assert.NoError(t, kv.Open())
	})
	t.Run(`configure`, func(t *testing.T) {
		kv.Configure(cluster.Config{
			NodeID:        "node",
			Advertise:     "127.0.0.1:7654",
			RetryInterval: time.Second,
			TTL:           time.Second * 30,
			BackendURL:    fmt.Sprintf("consul://%s/first", srv.Address()),
		})
	})
	t.Run(`store and subscribe before consul`, func(t *testing.T) {
		kv.Submit([]cluster.StoreOp{
			{
				Message: bus.NewMessage("up/v-1", map[string]interface{}{"V": 1}),
				WithTTL: true,
			},
			{
				Message: bus.NewMessage("up/p-1", map[string]interface{}{"P": 1}),
				WithTTL: true,
			},
		})
		//kv.Subscribe("down", watcherCtx, watcher)
	})
	t.Run(`start Consul`, func(t *testing.T) {
		srv.Up()
		srv.WaitLeader()
	})
	t.Run(`ensure stored messages in first`, func(t *testing.T) {
		fixture.WaitNoErrorT(t, waitConfig, func() (err error) {
			res, _, err := cli.KV().List("first/up", nil)
			if err != nil {
				return
			}
			if len(res) != 2 {
				err = fmt.Errorf(`expected two keys in first/up`)
			}
			return
		})
	})
	t.Run(`reconfigure with new node id`, func(t *testing.T) {
		kv.Configure(cluster.Config{
			NodeID:        "node-2",
			Advertise:     "127.0.0.1:7654",
			RetryInterval: time.Second,
			TTL:           time.Second * 30,
			BackendURL:    fmt.Sprintf("consul://%s/second", srv.Address()),
		})
	})
	t.Run(`ensure no messages in first`, func(t *testing.T) {
		wc2 := fixture.DefaultWaitConfig()
		wc2.Retries = 2
		fixture.WaitNoErrorT(t, wc2, func() (err error) {

			res, _, err := cli.KV().List("first/up", nil)
			if err != nil {
				return
			}
			if len(res) != 0 {
				err = fmt.Errorf(`expected no keys in first/up`)
			}
			return
		})
	})
	t.Run(`ensure stored messages in second`, func(t *testing.T) {
		fixture.WaitNoErrorT(t, waitConfig, func() (err error) {
			res, _, err := cli.KV().List("second/up", nil)
			if err != nil {
				return
			}
			if len(res) != 2 {
				err = fmt.Errorf(`expected two keys in second/up`)
			}
			return
		})
	})

	kv.Close()
	kv.Wait()
}
