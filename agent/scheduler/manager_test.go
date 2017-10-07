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

type testManagerEntity struct {
	name       string
	mu         sync.Mutex
	Count      int
	Reason     error
	Env        map[string]string
	Registered bool
}

func (e *testManagerEntity) notifyFn(reason error, message bus.Message) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Count++
	e.Reason = reason
	e.Env = message.GetPayload()
	e.Registered = true
}

func (e *testManagerEntity) unregisterFn() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Count++
	e.Registered = false
}

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

func TestManager_RegisterResource(t *testing.T) {
	ctx := context.Background()
	log := logx.GetLog("test")

	manager := scheduler.NewManager(ctx, log, "test",
		scheduler.NewManagerSource("meta", false, nil, "private", "public"),
		scheduler.NewManagerSource("with.dot", true, nil, "private", "public"),
		scheduler.NewManagerSource("drain", false, manifest.Constraint{
			"${drain.state}": "!= true",
		}, "private", "public"),
	)
	meta := bus.NewStrictMapUpstream("meta", manager)
	withDot := bus.NewStrictMapUpstream("with.dot", manager)
	drain := bus.NewStrictMapUpstream("drain", manager)

	sv := supervisor.NewChain(ctx, manager)
	assert.NoError(t, sv.Open())

	meta.Set(map[string]string{
		"first":  "1",
		"second": "1",
	})

	entity1 := &testManagerEntity{
		name: "1",
	}
	entity2 := &testManagerEntity{
		name: "2",
	}

	t.Run("0 register_first", func(t *testing.T) {
		manager.RegisterResource("first", manifest.PrivateNamespace, manifest.Constraint{
			"${meta.first}": "1",
		}, entity1.notifyFn)
		time.Sleep(time.Millisecond * 100)
		assert.Equal(t, entity1.Count, 0, "should not be notified (private is blocked)")
		assert.Nil(t, entity1.Env)
	})
	t.Run("1 enable with.dot", func(t *testing.T) {
		withDot.Set(map[string]string{
			"first":  "1",
			"second": "1",
		})
		time.Sleep(time.Millisecond * 100)
		assert.Equal(t, entity1.Count, 1)
		assert.Equal(t, entity2.Count, 0)

		assert.NotNil(t, entity1.Env)
		assert.NoError(t, entity1.Reason)
		assert.Equal(t, entity1.Env, map[string]string{
			"meta.first":  "1",
			"meta.second": "1",
		})
	})
	t.Run("2 register second", func(t *testing.T) {
		manager.RegisterResource("second", manifest.PrivateNamespace, manifest.Constraint{
			"${meta.second}":    "1",
			"${with.dot.first}": "1",
		}, entity2.notifyFn)
		time.Sleep(time.Millisecond * 100)

		assert.Equal(t, entity1.Count, 1, "first should not be notified")
		assert.Equal(t, entity2.Count, 1, "second should be notified")

		assert.NoError(t, entity2.Reason)
		assert.Equal(t, entity2.Env, map[string]string{
			"meta.first":  "1",
			"meta.second": "1",
		})
	})
	t.Run("drain on", func(t *testing.T) {
		drain.Set(map[string]string{
			"state": "true",
		})
		time.Sleep(time.Millisecond * 100)
		assert.Equal(t, entity1.Count, 2, "first should be notified")
		assert.Equal(t, entity2.Count, 2, "second should be notified")
		assert.Error(t, entity1.Reason)
		assert.Error(t, entity2.Reason)

	})
	t.Run("drain off", func(t *testing.T) {
		drain.Set(map[string]string{})
		time.Sleep(time.Millisecond * 100)

		assert.Equal(t, entity1.Count, 3, "first should be notified")
		assert.Equal(t, entity2.Count, 3, "second should be notified")
		assert.NoError(t, entity1.Reason)
		assert.NoError(t, entity2.Reason)
		assert.Equal(t, entity1.Env, map[string]string{
			"meta.first":  "1",
			"meta.second": "1",
		})
		assert.Equal(t, entity2.Env, map[string]string{
			"meta.first":  "1",
			"meta.second": "1",
		})
	})
	t.Run("fail second constraint", func(t *testing.T) {
		meta.Set(map[string]string{
			"first": "1",
		})
		time.Sleep(time.Millisecond * 100)
		assert.Equal(t, entity1.Count, 4, "first should be notified")
		assert.Equal(t, entity2.Count, 4, "second should be notified")
		assert.NoError(t, entity1.Reason)
		assert.Error(t, entity2.Reason)
	})
	t.Run("reregister first", func(t *testing.T) {
		manager.RegisterResource("first", manifest.PrivateNamespace, manifest.Constraint{
			"${meta.first}": "1",
		}, entity1.notifyFn)
		time.Sleep(time.Millisecond * 100)
		assert.Equal(t, entity1.Count, 5, "first should be notified")
		assert.True(t, entity1.Registered, "first should be registered")
		assert.Equal(t, entity2.Count, 4, "second should not be notified")
	})
	t.Run("deregister first", func(t *testing.T) {
		manager.UnregisterResource("first", entity1.unregisterFn)
		time.Sleep(time.Millisecond * 100)
		assert.Equal(t, entity1.Count, 6, "first should be notified")
		assert.False(t, entity1.Registered, "first should be unregistered")
		assert.Equal(t, entity2.Count, 4, "second not should be notified")
		assert.NoError(t, entity1.Reason)
		assert.Error(t, entity2.Reason)

		meta.Set(map[string]string{
			"first":  "2",
			"second": "1",
		})
		time.Sleep(time.Millisecond * 100)
		assert.Equal(t, entity1.Count, 6, "first not should be notified")
		assert.Equal(t, entity2.Count, 5, "second should be notified")
		assert.NoError(t, entity1.Reason)
		assert.NoError(t, entity2.Reason)
	})

	assert.NoError(t, sv.Close())
	assert.NoError(t, sv.Wait())
}

func TestManager_OverlapSources(t *testing.T) {
	ctx := context.Background()
	log := logx.GetLog("test")

	manager := scheduler.NewManager(ctx, log, "test",
		scheduler.NewManagerSource("meta", false, nil, manifest.PrivateNamespace, manifest.PublicNamespace),
		scheduler.NewManagerSource("meta.private", false, nil, manifest.PrivateNamespace),
		scheduler.NewManagerSource("meta.public", false, nil, manifest.PublicNamespace),
	)
	metaMap := bus.NewStrictMapUpstream("meta", manager)
	privateMap := bus.NewStrictMapUpstream("meta.private", manager)
	publicMap := bus.NewStrictMapUpstream("meta.public", manager)

	sv := supervisor.NewChain(ctx, manager)
	assert.NoError(t, sv.Open())

	entityMetaOnly := &testManagerEntity{}
	entityPrivateOnly := &testManagerEntity{}
	entityPublicOnly := &testManagerEntity{}
	entityMetaPrivate := &testManagerEntity{}
	entityMetaPublic := &testManagerEntity{}

	manager.RegisterResource("meta-only", manifest.PrivateNamespace, manifest.Constraint{
		"${meta.value}": "true",
	}, entityMetaOnly.notifyFn)
	manager.RegisterResource("private-only", manifest.PrivateNamespace, manifest.Constraint{
		"${meta.private.value}": "true",
	}, entityPrivateOnly.notifyFn)
	manager.RegisterResource("public-only", manifest.PublicNamespace, manifest.Constraint{
		"${meta.public.value}": "true",
	}, entityPublicOnly.notifyFn)
	manager.RegisterResource("meta-private", manifest.PrivateNamespace, manifest.Constraint{
		"${meta.value}":        "true",
		"${meta.public.value}": "true",
	}, entityMetaPrivate.notifyFn)
	manager.RegisterResource("meta-public", manifest.PublicNamespace, manifest.Constraint{
		"${meta.value}":        "true",
		"${meta.public.value}": "true",
	}, entityMetaPublic.notifyFn)

	t.Run("0 check all inactive", func(t *testing.T) {
		time.Sleep(time.Millisecond * 100)
		assert.Equal(t, []int{0, 0, 0, 0, 0}, []int{
			entityMetaOnly.Count,
			entityPrivateOnly.Count,
			entityPublicOnly.Count,
			entityMetaPrivate.Count,
			entityMetaPublic.Count,
		}, "all should are not notified")
	})
	t.Run("1 meta.private.value=true", func(t *testing.T) {
		privateMap.Set(map[string]string{
			"value": "true",
		})
		time.Sleep(time.Millisecond * 100)
		assert.Equal(t, []int{0, 0, 0, 0, 0}, []int{
			entityMetaOnly.Count,
			entityPrivateOnly.Count,
			entityPublicOnly.Count,
			entityMetaPrivate.Count,
			entityMetaPublic.Count,
		}, "all should are not notified")
	})
	t.Run("2 meta.value=true", func(t *testing.T) {
		metaMap.Set(map[string]string{
			"value": "true",
		})
		time.Sleep(time.Millisecond * 100)
		assert.Equal(t, []int{1, 1, 0, 1, 0}, []int{
			entityMetaOnly.Count,
			entityPrivateOnly.Count,
			entityPublicOnly.Count,
			entityMetaPrivate.Count,
			entityMetaPublic.Count,
		}, "only private should be notified")
	})
	t.Run("3 meta.public.value=true", func(t *testing.T) {
		publicMap.Set(map[string]string{
			"value": "true",
		})
		time.Sleep(time.Millisecond * 100)
		assert.Equal(t, []int{2, 2, 1, 2, 1}, []int{
			entityMetaOnly.Count,
			entityPrivateOnly.Count,
			entityPublicOnly.Count,
			entityMetaPrivate.Count,
			entityMetaPublic.Count,
		}, "all should be notified")
	})

	assert.NoError(t, sv.Close())
	assert.NoError(t, sv.Wait())
}
