// +build ide test_cluster_consul

package cluster_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/cluster"
	"github.com/akaspin/soil/fixture"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewConsulWorker(t *testing.T) {
	t.Skip()
	srv := fixture.NewConsulServer(t, nil)
	defer srv.Clean()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log := logx.GetLog("test")

	cli, err := api.NewClient(&api.Config{
		Address: srv.Address(),
	})
	assert.NoError(t, err)

	t.Run(`no server`, func(t *testing.T) {
		worker := cluster.NewConsulWorker(ctx, log, cluster.WorkerConfig{
			Address: srv.Address(),
		})
		select {
		case <-worker.CleanCtx().Done():
			t.Error(`should be not clean`)
			t.Fail()
		case <-worker.FailureCtx().Done():
		}
	})
	t.Run(`up`, func(t *testing.T) {
		srv.Up()
		srv.WaitAlive()
	})
	t.Run(`worker 1`, func(t *testing.T) {
		worker := cluster.NewConsulWorker(ctx, log, cluster.WorkerConfig{
			Address: srv.Address(),
			TTL:     time.Minute,
			ID:      "test-node",
		})
		defer worker.Close()
		select {
		case <-worker.CleanCtx().Done():
		case <-worker.FailureCtx().Done():
			t.Error(`should not fail`)
			t.Fail()
		}
		sessions, _, err := cli.Session().List(nil)
		assert.NoError(t, err)
		assert.Len(t, sessions, 1)
		assert.Equal(t, "test-node", sessions[0].Name)
	})
	t.Run(`worker 2`, func(t *testing.T) {
		worker := cluster.NewConsulWorker(ctx, log, cluster.WorkerConfig{
			Address: srv.Address(),
			TTL:     time.Minute,
			ID:      "test-node",
		})
		defer worker.Close()
		select {
		case <-worker.CleanCtx().Done():
		case <-worker.FailureCtx().Done():
			t.Error(`should not fail`)
			t.Fail()
		}
		sessions, _, err := cli.Session().List(nil)
		assert.NoError(t, err)
		assert.Len(t, sessions, 1)
		assert.Equal(t, "test-node", sessions[0].Name)
	})
}

func TestConsulWorker_Submit(t *testing.T) {
	srv := fixture.NewConsulServer(t, nil)
	defer srv.Clean()
	srv.Up()
	srv.WaitAlive()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log := logx.GetLog("test")
	cli, cliErr := api.NewClient(&api.Config{
		Address: srv.Address(),
	})
	assert.NoError(t, cliErr)

	worker := cluster.NewConsulWorker(ctx, log, cluster.WorkerConfig{
		Address: srv.Address(),
		TTL:     time.Second * 2,
	})
	defer worker.Close()

	var commits []string
	go func() {
		for keys := range worker.CommitChan() {
			commits = append(commits, keys...)
		}
	}()

	t.Run("wait clean", func(t *testing.T) {
		select {
		case <-worker.CleanCtx().Done():
		case <-worker.FailureCtx().Done():
			t.Error(`should not fail`)
			t.Fail()
		}
	})
	t.Run("submit", func(t *testing.T) {
		worker.Submit([]cluster.WorkerStoreOp{
			{
				Message: bus.NewMessage("test/01", "01"),
			},
			{
				Message: bus.NewMessage("test/02", "02"),
				WithTTL: true,
			},
		})
		time.Sleep(time.Millisecond * 300)
		assert.Equal(t, commits, []string{"test/01", "test/02"})
	})
	t.Run("ensure", func(t *testing.T) {
		res, _, err := cli.KV().List("test/", nil)
		assert.NoError(t, err)
		assert.Len(t, res, 2)
	})
	t.Run("ensure ttl", func(t *testing.T) {
		time.Sleep(time.Second * 2)
		res, _, err := cli.KV().Get("test/02", nil)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, []byte(`"02"`), res.Value)
	})
	t.Run("submit delete", func(t *testing.T) {
		worker.Submit([]cluster.WorkerStoreOp{
			{
				Message: bus.NewMessage("test/01", nil),
			},
			{
				Message: bus.NewMessage("test/02", nil),
			},
			{
				Message: bus.NewMessage("test/03", nil),
			},
		})
		time.Sleep(time.Millisecond * 300)
		assert.Equal(t, commits, []string{
			"test/01", "test/02",
			"test/01", "test/02", "test/03",
		})
		res, _, err := cli.KV().List("test/", nil)
		assert.NoError(t, err)
		assert.Len(t, res, 0)
	})
}
