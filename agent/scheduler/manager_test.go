// +build ide test_unit

package scheduler_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestMap_Nil(t *testing.T) {
	a := func(data map[string]string, constraint manifest.Constraint) {
		if data != nil {
			t.Fail()
		}
		if constraint != nil {
			t.Fail()
		}
	}
	a(nil, nil)
}

func TestManager(t *testing.T) {
	ctx := context.Background()
	log := logx.GetLog("test")

	manager := scheduler.NewManager(ctx, log,
		scheduler.NewManagerSource("meta", false, nil, "private", "public"),
		scheduler.NewManagerSource("with.dot", true, nil, "private", "public"),
		scheduler.NewManagerSource("drain", false, manifest.Constraint{
			"${drain.state}": "!= true",
		}, "private", "public"),
	)
	a1 := bus.NewFlatMap(ctx, log, true, "meta", manager)
	a2 := bus.NewFlatMap(ctx, log, true, "with.dot", manager)
	drainMeta := bus.NewFlatMap(ctx, log, true, "drain", manager)

	sv := supervisor.NewChain(ctx, manager, a1, a2, drainMeta)
	assert.NoError(t, sv.Open())

	a1.Set(map[string]string{
		"first":  "1",
		"second": "1",
	})
	a2.Set(map[string]string{
		"first":  "1",
		"second": "1",
	})

	var privatePods manifest.Registry
	err := privatePods.UnmarshalFiles("private", "testdata/manager_test.hcl")
	assert.NoError(t, err)

	mu := &sync.Mutex{}
	type Res struct {
		Reason error
		Env    map[string]string
		Mark   uint64
	}
	res := map[string][]Res{}

	handler := func(n string, reason error, env map[string]string, mark uint64) {
		mu.Lock()
		defer mu.Unlock()
		res[n] = append(res[n], Res{
			Reason: reason,
			Env:    env,
			Mark:   mark,
		})
	}

	t.Run("register first", func(t *testing.T) {
		manager.RegisterResource("first", privatePods[0].Namespace, privatePods[0].Constraint, func(reason error, env map[string]string, mark uint64) {
			handler("first", reason, env, mark)
		})
		time.Sleep(time.Millisecond * 100)
		assert.Len(t, res["first"], 1, "first should be notified")
		assert.NoError(t, res["first"][0].Reason)
		assert.Equal(t, res["first"][0].Env, map[string]string{
			"meta.first":  "1",
			"meta.second": "1",
		})

	})
	t.Run("register second", func(t *testing.T) {
		manager.RegisterResource("second", privatePods[1].Namespace, privatePods[1].Constraint, func(reason error, env map[string]string, mark uint64) {
			handler("second", reason, env, mark)
		})
		time.Sleep(time.Millisecond * 100)

		assert.Len(t, res["first"], 1, "first should not be notified")
		assert.Len(t, res["second"], 1, "second should be notified")

		assert.NoError(t, res["second"][0].Reason)
		assert.Equal(t, res["second"][0].Env, map[string]string{
			"meta.first":  "1",
			"meta.second": "1",
		})
	})
	t.Run("drain on", func(t *testing.T) {
		drainMeta.Set(map[string]string{
			"state": "true",
		})
		time.Sleep(time.Millisecond * 100)
		assert.Len(t, res["first"], 2, "first should be notified")
		assert.Len(t, res["second"], 2, "second should be notified")
		assert.Error(t, res["first"][1].Reason)
		assert.Error(t, res["second"][1].Reason)
	})
	t.Run("drain off", func(t *testing.T) {
		drainMeta.Set(map[string]string{})
		time.Sleep(time.Millisecond * 100)
		assert.Len(t, res["first"], 3, "first should be notified")
		assert.Len(t, res["second"], 3, "second should be notified")
		assert.NoError(t, res["first"][2].Reason)
		assert.NoError(t, res["second"][2].Reason)
		assert.Equal(t, res["first"][2].Env, map[string]string{
			"meta.first":  "1",
			"meta.second": "1",
		})
		assert.Equal(t, res["second"][2].Env, map[string]string{
			"meta.first":  "1",
			"meta.second": "1",
		})
	})
	t.Run("fail second constraint", func(t *testing.T) {
		a1.Set(map[string]string{
			"first": "1",
		})
		time.Sleep(time.Millisecond * 100)
		assert.Len(t, res["first"], 4, "first should be notified")
		assert.Len(t, res["second"], 4, "second should be notified")
		//pretty.Log(res["first"])
		assert.NoError(t, res["first"][3].Reason)
		assert.Error(t, res["second"][3].Reason)
	})

	assert.NoError(t, sv.Close())
	assert.NoError(t, sv.Wait())

}
