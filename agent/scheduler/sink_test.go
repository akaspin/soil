// +build ide test_unit

package scheduler_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/lib"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"github.com/mitchellh/hashstructure"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

type dummyEvRecord struct {
	alloc bool
	pod   uint64
	env   uint64
}

type dummyEv struct {
	mu      sync.Mutex
	records map[string][]dummyEvRecord
}

func (e *dummyEv) GetConstraint(pod *manifest.Pod) manifest.Constraint {
	return pod.GetResourceAllocationConstraint()
}

func (e *dummyEv) Allocate(pod *manifest.Pod, env map[string]string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	record := dummyEvRecord{
		alloc: true,
		pod:   pod.Mark(),
	}
	record.env, _ = hashstructure.Hash(env, nil)
	if e.records == nil {
		e.records = map[string][]dummyEvRecord{}
	}
	e.records[pod.Name] = append(e.records[pod.Name], record)
}

func (e *dummyEv) Deallocate(name string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.records == nil {
		e.records = map[string][]dummyEvRecord{}
	}
	e.records[name] = append(e.records[name], dummyEvRecord{})
}

func TestSink_ConsumeRegistry(t *testing.T) {
	ctx := context.Background()
	log := logx.GetLog("test")

	arbiter1 := scheduler.NewArbiter(ctx, log, "a1",
		scheduler.ArbiterConfig{
			Required: manifest.Constraint{
				"${drain}": "!= true",
			},
		},
	)
	evaluator1 := &dummyEv{}
	sink := scheduler.NewSink(ctx, log, nil, scheduler.NewBoundedEvaluator(arbiter1, evaluator1))
	sv := supervisor.NewChain(ctx, arbiter1, sink)
	assert.NoError(t, sv.Open())

	time.Sleep(time.Millisecond * 100)

	t.Run("0 consume", func(t *testing.T) {
		var buffers lib.StaticBuffers
		var registry manifest.Pods
		assert.NoError(t, buffers.ReadFiles("testdata/sink_test_ConsumeRegistry_0.hcl"))
		assert.NoError(t, registry.Unmarshal(manifest.PrivateNamespace, buffers.GetReaders()...))
		sink.ConsumeRegistry(registry)
		time.Sleep(time.Millisecond * 100)

		assert.Nil(t, evaluator1.records, "no allocations")
	})
	t.Run("1 enable first and second", func(t *testing.T) {
		arbiter1.ConsumeMessage(bus.NewMessage("", map[string]string{
			"meta.first":  "true",
			"meta.second": "true",
		}))
		time.Sleep(time.Millisecond * 100)
		assert.Equal(t,
			map[string][]dummyEvRecord{
				"first":  {{alloc: true, pod: 0xcc3091b7b018e0dc, env: 0x88be0fba4063a209}},
				"second": {{alloc: true, pod: 0xc089f45ab17a8664, env: 0x88be0fba4063a209}},
				"third":  {{alloc: false, pod: 0x0, env: 0x0}},
			},
			evaluator1.records)
	})
	t.Run("2 modify third", func(t *testing.T) {
		var buffers lib.StaticBuffers
		var registry manifest.Pods
		assert.NoError(t, buffers.ReadFiles("testdata/sink_test_ConsumeRegistry_2.hcl"))
		assert.NoError(t, registry.Unmarshal(manifest.PrivateNamespace, buffers.GetReaders()...))
		sink.ConsumeRegistry(registry)
		time.Sleep(time.Millisecond * 100)

		assert.Equal(t, map[string][]dummyEvRecord{
			"first":  {{alloc: true, pod: 0xcc3091b7b018e0dc, env: 0x88be0fba4063a209}},
			"second": {{alloc: true, pod: 0xc089f45ab17a8664, env: 0x88be0fba4063a209}},
			"third":  {{alloc: false, pod: 0x0, env: 0x0}, {alloc: true, pod: 0x476e164f1b6945e9, env: 0x88be0fba4063a209}},
		},
			evaluator1.records, "third should be updated")
	})
	t.Run("3 deactivate", func(t *testing.T) {
		arbiter1.ConsumeMessage(bus.NewMessage("", nil))
		time.Sleep(time.Millisecond * 100)
		assert.Equal(t, map[string][]dummyEvRecord{
			"first":  {{alloc: true, pod: 0xcc3091b7b018e0dc, env: 0x88be0fba4063a209}},
			"second": {{alloc: true, pod: 0xc089f45ab17a8664, env: 0x88be0fba4063a209}},
			"third":  {{alloc: false, pod: 0x0, env: 0x0}, {alloc: true, pod: 0x476e164f1b6945e9, env: 0x88be0fba4063a209}},
		}, evaluator1.records, "no updates: inactive")
	})
	t.Run("4 remove third", func(t *testing.T) {
		var buffers lib.StaticBuffers
		var registry manifest.Pods
		assert.NoError(t, buffers.ReadFiles("testdata/sink_test_ConsumeRegistry_4.hcl"))
		assert.NoError(t, registry.Unmarshal(manifest.PrivateNamespace, buffers.GetReaders()...))
		sink.ConsumeRegistry(registry)
		time.Sleep(time.Millisecond * 100)

		assert.Equal(t, map[string][]dummyEvRecord{
			"second": {{alloc: true, pod: 0xc089f45ab17a8664, env: 0x88be0fba4063a209}},
			"third":  {{alloc: false, pod: 0x0, env: 0x0}, {alloc: true, pod: 0x476e164f1b6945e9, env: 0x88be0fba4063a209}, {alloc: false, pod: 0x0, env: 0x0}},
			"first":  {{alloc: true, pod: 0xcc3091b7b018e0dc, env: 0x88be0fba4063a209}},
		}, evaluator1.records, "no updates: inactive")
	})
	t.Run("5 activate", func(t *testing.T) {
		arbiter1.ConsumeMessage(bus.NewMessage("", map[string]string{
			"meta.first":  "true",
			"meta.second": "true",
		}))
		time.Sleep(time.Millisecond * 100)
		assert.Equal(t, map[string][]dummyEvRecord{
			"first":  {{alloc: true, pod: 0xcc3091b7b018e0dc, env: 0x88be0fba4063a209}, {alloc: true, pod: 0xcc3091b7b018e0dc, env: 0x88be0fba4063a209}},
			"second": {{alloc: true, pod: 0xc089f45ab17a8664, env: 0x88be0fba4063a209}, {alloc: true, pod: 0xc089f45ab17a8664, env: 0x88be0fba4063a209}},
			"third":  {{alloc: false, pod: 0x0, env: 0x0}, {alloc: true, pod: 0x476e164f1b6945e9, env: 0x88be0fba4063a209}, {alloc: false, pod: 0x0, env: 0x0}},
		}, evaluator1.records)

	})
	t.Run("6 drain", func(t *testing.T) {
		arbiter1.ConsumeMessage(bus.NewMessage("", map[string]string{
			"drain": "true",
		}))
		time.Sleep(time.Millisecond * 100)
		assert.Equal(t, map[string][]dummyEvRecord{
			"first":  {{alloc: true, pod: 0xcc3091b7b018e0dc, env: 0x88be0fba4063a209}, {alloc: true, pod: 0xcc3091b7b018e0dc, env: 0x88be0fba4063a209}, {alloc: false, pod: 0x0, env: 0x0}},
			"second": {{alloc: true, pod: 0xc089f45ab17a8664, env: 0x88be0fba4063a209}, {alloc: true, pod: 0xc089f45ab17a8664, env: 0x88be0fba4063a209}, {alloc: false, pod: 0x0, env: 0x0}},
			"third":  {{alloc: false, pod: 0x0, env: 0x0}, {alloc: true, pod: 0x476e164f1b6945e9, env: 0x88be0fba4063a209}, {alloc: false, pod: 0x0, env: 0x0}},
		}, evaluator1.records, "drain")
	})
	t.Run("7 remove drain", func(t *testing.T) {
		arbiter1.ConsumeMessage(bus.NewMessage("", map[string]string{
			"meta.first":  "true",
			"meta.second": "true",
		}))
		time.Sleep(time.Millisecond * 100)
		assert.Equal(t, map[string][]dummyEvRecord{
			"first":  {{alloc: true, pod: 0xcc3091b7b018e0dc, env: 0x88be0fba4063a209}, {alloc: true, pod: 0xcc3091b7b018e0dc, env: 0x88be0fba4063a209}, {alloc: false, pod: 0x0, env: 0x0}, {alloc: true, pod: 0xcc3091b7b018e0dc, env: 0x88be0fba4063a209}},
			"second": {{alloc: true, pod: 0xc089f45ab17a8664, env: 0x88be0fba4063a209}, {alloc: true, pod: 0xc089f45ab17a8664, env: 0x88be0fba4063a209}, {alloc: false, pod: 0x0, env: 0x0}, {alloc: true, pod: 0xc089f45ab17a8664, env: 0x88be0fba4063a209}},
			"third":  {{alloc: false, pod: 0x0, env: 0x0}, {alloc: true, pod: 0x476e164f1b6945e9, env: 0x88be0fba4063a209}, {alloc: false, pod: 0x0, env: 0x0}},
		}, evaluator1.records, "remove drain")
	})

	sv.Close()
	sv.Wait()
}
