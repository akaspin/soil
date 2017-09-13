// +build ide test_unit

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

	a1 := source.NewPlain(ctx, log, "meta", true)
	a2 := source.NewPlain(ctx, log, "with.dot", false)

	man := scheduler.NewArbiter(ctx, log, a1, a2)
	sv := supervisor.NewChain(ctx, a1, man)
	assert.NoError(t, sv.Open())

	a1.Set(map[string]string{
		"first":  "1",
		"second": "1",
	}, true)
	a2.Set(map[string]string{
		"first":  "1",
		"second": "1",
	}, true)

	privatePods, err := manifest.ParseFromFiles("private", "testdata/arbiter_test.hcl")
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
	t.Run("drain on", func(t *testing.T) {
		man.Drain(true)
		time.Sleep(time.Millisecond * 100)
		assert.Error(t, res["first"])
		assert.Error(t, res["second"])
	})
	t.Run("drain off", func(t *testing.T) {
		man.Drain(false)
		time.Sleep(time.Millisecond * 100)
		assert.NoError(t, res["first"])
		assert.NoError(t, res["second"])
	})
	t.Run("off second", func(t *testing.T) {
		a1.Set(map[string]string{
			"first": "1",
		}, true)
		time.Sleep(time.Millisecond * 100)
		assert.NoError(t, res["first"])
		assert.Error(t, res["second"])
	})

	assert.NoError(t, sv.Close())
	assert.NoError(t, sv.Wait())

}
