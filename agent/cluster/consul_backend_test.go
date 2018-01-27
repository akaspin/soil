// +build ide test_cluster

package cluster_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/cluster"
	"github.com/akaspin/soil/fixture"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

func TestNewConsulBackend(t *testing.T) {
	srv := fixture.NewConsulServer(t, nil)
	defer srv.Clean()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log := logx.GetLog("test")

	cli, cliErr := api.NewClient(&api.Config{
		Address: srv.Address(),
	})
	assert.NoError(t, cliErr)

	t.Run(`no server`, func(t *testing.T) {
		kv := cluster.NewConsulBackend(ctx, log, cluster.BackendConfig{
			Address: srv.Address(),
		})
		select {
		case <-kv.ReadyCtx().Done():
			t.Error(`should be not clean`)
			t.Fail()
		case <-kv.FailCtx().Done():
		}
	})
	t.Run(`up`, func(t *testing.T) {
		srv.Up()
		srv.WaitLeader()
	})
	t.Run(`kv 1`, func(t *testing.T) {
		kv := cluster.NewConsulBackend(ctx, log, cluster.BackendConfig{
			Address: srv.Address(),
			TTL:     time.Minute,
			ID:      "test-node",
		})
		defer kv.Close()
		select {
		case <-kv.ReadyCtx().Done():
		case <-kv.FailCtx().Done():
			t.Error(`should not fail`)
			t.Fail()
		}
		sessions, _, err := cli.Session().List(nil)
		assert.NoError(t, err)
		require.NotNil(t, sessions)
		assert.Len(t, sessions, 1)
		assert.Equal(t, "test-node", sessions[0].Name)
	})
	t.Run(`kv 2`, func(t *testing.T) {
		kv := cluster.NewConsulBackend(ctx, log, cluster.BackendConfig{
			Address: srv.Address(),
			TTL:     time.Minute,
			ID:      "test-node",
		})
		defer kv.Close()
		select {
		case <-kv.ReadyCtx().Done():
		case <-kv.FailCtx().Done():
			t.Error(`should not fail`)
			t.Fail()
		}
		sessions, _, err := cli.Session().List(nil)
		assert.NoError(t, err)
		require.NotNil(t, sessions)
		assert.Len(t, sessions, 1)
		assert.Equal(t, "test-node", sessions[0].Name)
	})
}

func TestConsulBackend_Submit(t *testing.T) {
	srv := fixture.NewConsulServer(t, nil)
	defer srv.Clean()
	srv.Up()
	srv.WaitLeader()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log := logx.GetLog("test")
	cli, cliErr := api.NewClient(&api.Config{
		Address: srv.Address(),
	})
	assert.NoError(t, cliErr)

	kv := cluster.NewConsulBackend(ctx, log, cluster.BackendConfig{
		Address: srv.Address(),
		TTL:     time.Second * 2,
		Chroot:  "soil",
		ID:      "node",
	})
	defer kv.Close()

	commitsMu := &sync.Mutex{}
	var commits []cluster.StoreCommit
	go func() {
		for ok := range kv.CommitChan() {
			commitsMu.Lock()
			commits = append(commits, ok...)
			commitsMu.Unlock()
		}
	}()

	t.Run("wait clean", func(t *testing.T) {
		select {
		case <-kv.ReadyCtx().Done():
		case <-kv.FailCtx().Done():
			t.Error(`should not fail`)
			t.Fail()
		}
	})
	t.Run("submit", func(t *testing.T) {
		kv.Submit([]cluster.StoreOp{
			{
				Message: bus.NewMessage("test/01", "01"),
			},
			{
				Message: bus.NewMessage("test/02", "02"),
				WithTTL: true,
			},
		})
		time.Sleep(time.Millisecond * 300)
		commitsMu.Lock()
		assert.Equal(t, commits, []cluster.StoreCommit{
			{ID: "test/01", Hash: 0x814776e2108083a4, WithTTL: false},
			{ID: "test/02", Hash: 0x7c7cfc54f5f190b3, WithTTL: true}})
		commitsMu.Unlock()
	})
	t.Run("ensure", func(t *testing.T) {
		res, _, err := cli.KV().List("soil/test/", nil)
		assert.NoError(t, err)
		assert.Len(t, res, 2)
	})
	t.Run("ensure volatile", func(t *testing.T) {
		time.Sleep(time.Second * 2)
		res, _, err := cli.KV().Get("soil/test/02/node", nil)
		assert.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, []byte(`"02"`), res.Value)
	})
	t.Run("submit delete", func(t *testing.T) {
		kv.Submit([]cluster.StoreOp{
			{
				Message: bus.NewMessage("test/01", nil),
			},
			{
				Message: bus.NewMessage("test/02", nil),
				WithTTL: true,
			},
			{
				Message: bus.NewMessage("test/03", nil),
			},
		})
		time.Sleep(time.Millisecond * 300)
		commitsMu.Lock()
		assert.Equal(t, commits, []cluster.StoreCommit{
			{ID: "test/01", Hash: 0x814776e2108083a4, WithTTL: false},
			{ID: "test/02", Hash: 0x7c7cfc54f5f190b3, WithTTL: true},
			{ID: "test/01", Hash: 0x0, WithTTL: false},
			{ID: "test/02", Hash: 0x0, WithTTL: true},
			{ID: "test/03", Hash: 0x0, WithTTL: false}})
		commitsMu.Unlock()
		res, _, err := cli.KV().List("soil/test/", nil)
		assert.NoError(t, err)
		assert.Len(t, res, 0)
	})
}

func TestConsulBackend_Subscribe(t *testing.T) {
	srv := fixture.NewConsulServer(t, nil)
	defer srv.Clean()
	srv.Up()
	srv.WaitLeader()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log := logx.GetLog("test")
	cli, cliErr := api.NewClient(&api.Config{
		Address: srv.Address(),
	})
	assert.NoError(t, cliErr)

	kv := cluster.NewConsulBackend(ctx, log, cluster.BackendConfig{
		Address: srv.Address(),
		TTL:     time.Second * 2,
		Chroot:  "soil",
	})
	defer kv.Close()

	cons1 := bus.NewTestingConsumer(ctx)

	go func() {
		for result := range kv.WatchResultsChan() {
			payload := map[string]interface{}{}
			for k, v := range result.Data {
				var value interface{}
				assert.NoError(t, json.NewDecoder(bytes.NewReader(v)).Decode(&value))
				payload[k] = value
			}
			cons1.ConsumeMessage(bus.NewMessage(result.Key, payload))
		}
	}()

	watch1Ctx, watch1Cancel := context.WithCancel(ctx)
	watch2Ctx, watch2Cancel := context.WithCancel(ctx)
	defer watch1Cancel()
	defer watch2Cancel()

	t.Run("wait clean", func(t *testing.T) {
		select {
		case <-kv.ReadyCtx().Done():
		case <-kv.Ctx().Done():
			t.Error(`should not fail`)
			t.Fail()
		}
	})
	t.Run(`setup watches`, func(t *testing.T) {
		kv.Subscribe([]cluster.WatchRequest{
			{
				Key: "1",
				Ctx: watch1Ctx,
			},
			{
				Key: "2",
				Ctx: watch2Ctx,
			},
		})
	})
	t.Run(`ensure empty messages`, func(t *testing.T) {
		fixture.WaitNoError(t, fixture.DefaultWaitConfig(), cons1.ExpectMessagesByIdFn(map[string][]bus.Message{
			"1": {
				bus.NewMessage("1", map[string]interface{}{}),
			},
			"2": {
				bus.NewMessage("2", map[string]interface{}{}),
			},
		}))
	})
	t.Run(`put to 1`, func(t *testing.T) {
		_, err := cli.KV().Put(&api.KVPair{
			Key:   "soil/1/one",
			Value: []byte(`"1"`),
		}, nil)
		assert.NoError(t, err)
	})
	t.Run(`ensure messages`, func(t *testing.T) {
		fixture.WaitNoError(t, fixture.DefaultWaitConfig(), cons1.ExpectMessagesByIdFn(map[string][]bus.Message{
			"1": {
				bus.NewMessage("1", map[string]interface{}{}),
				bus.NewMessage("1", map[string]interface{}{
					"one": "1",
				}),
			},
			"2": {
				bus.NewMessage("2", map[string]interface{}{}),
				bus.NewMessage("2", map[string]interface{}{}),
			},
		}))
	})
}
