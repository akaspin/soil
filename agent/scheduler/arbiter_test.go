package scheduler_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/agent/source"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestArbiter(t *testing.T) {
	ctx := context.Background()
	log := logx.GetLog("test")

	a1 := source.NewMapSource(ctx, log, "meta", true, manifest.Constraint{})
	a2 := source.NewMapSource(ctx, log, "agent", false, manifest.Constraint{})

	man := scheduler.NewArbiter(ctx, log, a1, a2)
	sv := supervisor.NewChain(ctx, a1, man)
	assert.NoError(t, sv.Open())

	a1.Configure(map[string]string{
		"first":  "1",
		"second": "1",
	})
	a2.Configure(map[string]string{
		"first":  "1",
		"second": "1",
	})

	privatePods, err := manifest.ParseFromFiles("private", "testdata/manager_test.hcl")
	assert.NoError(t, err)

	res := map[string]error{}
	mu := &sync.Mutex{}
	handler := func(n string, reason error, env map[string]string) {
		mu.Lock()
		defer mu.Unlock()
		res[n] = reason
	}

	t.Run("register first", func(t *testing.T) {
		man.Register("first", privatePods[0], func(reason error, env map[string]string, mark uint64) {
			handler("first", reason, env)
		})
		time.Sleep(time.Millisecond * 100)
		assert.NoError(t, res["first"])
	})
	t.Run("register second", func(t *testing.T) {
		man.Register("second", privatePods[1], func(reason error, env map[string]string, mark uint64) {
			handler("second", reason, env)
		})
		time.Sleep(time.Millisecond * 100)
		assert.NoError(t, res["first"])
		assert.NoError(t, res["second"])
	})
	t.Run("off second", func(t *testing.T) {
		a1.Configure(map[string]string{
			"first": "1",
		})
		time.Sleep(time.Millisecond * 100)
		assert.NoError(t, res["first"])
		assert.Error(t, res["second"])
	})

	assert.NoError(t, sv.Close())
	assert.NoError(t, sv.Wait())

}
