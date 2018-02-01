// +build ide test_unit

package scheduler_test

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/fixture"
	"github.com/akaspin/soil/lib"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"github.com/mitchellh/hashstructure"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strconv"
	"sync"
	"testing"
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

func (e *dummyEv) GetConstraint(pod *manifest.Pod) (res manifest.Constraint) {
	res = pod.Constraint.Clone()
	if len(pod.Resources) > 0 {
		c1 := manifest.Constraint{}
		for _, r := range pod.Resources {
			c1[fmt.Sprintf(`${resource.%s.%s.allocated}`, pod.Name, r.Name)] = "true"
		}
		res = res.Merge(c1)
	}
	return
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

func (e *dummyEv) assertFn(recs map[string][]dummyEvRecord) func() (err error) {
	return func() (err error) {
		e.mu.Lock()
		defer e.mu.Unlock()
		if !reflect.DeepEqual(recs, e.records) {
			err = fmt.Errorf("%v != %v", recs, e.records)
		}
		return
	}
}

func TestSink_ConsumeRegistry(t *testing.T) {
	for i := 0; i < 5; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
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

			t.Run("0 consume", func(t *testing.T) {
				var buffers lib.StaticBuffers
				var registry manifest.PodSlice
				assert.NoError(t, buffers.ReadFiles("testdata/sink_test_ConsumeRegistry_0.hcl"))
				assert.NoError(t, registry.Unmarshal(manifest.PrivateNamespace, buffers.GetReaders()...))
				sink.ConsumeRegistry(registry)
				fixture.WaitNoErrorT10(t, evaluator1.assertFn(nil))
			})
			t.Run("1 enable first and second", func(t *testing.T) {
				arbiter1.ConsumeMessage(bus.NewMessage("", map[string]string{
					"meta.first":  "true",
					"meta.second": "true",
				}))
				fixture.WaitNoErrorT10(t, evaluator1.assertFn(map[string][]dummyEvRecord{
					"first":  {{alloc: true, pod: 0x1c2120eda6a20a7f, env: 0x88be0fba4063a209}},
					"second": {{alloc: true, pod: 0xadbac957c6a4641f, env: 0x88be0fba4063a209}},
					"third":  {{alloc: false, pod: 0x0, env: 0x0}},
				}))
			})
			t.Run("2 modify third", func(t *testing.T) {
				var buffers lib.StaticBuffers
				var registry manifest.PodSlice
				assert.NoError(t, buffers.ReadFiles("testdata/sink_test_ConsumeRegistry_2.hcl"))
				assert.NoError(t, registry.Unmarshal(manifest.PrivateNamespace, buffers.GetReaders()...))
				sink.ConsumeRegistry(registry)
				fixture.WaitNoErrorT10(t, evaluator1.assertFn(map[string][]dummyEvRecord{
					"first":  {{alloc: true, pod: 0x1c2120eda6a20a7f, env: 0x88be0fba4063a209}},
					"second": {{alloc: true, pod: 0xadbac957c6a4641f, env: 0x88be0fba4063a209}},
					"third":  {{alloc: false, pod: 0x0, env: 0x0}, {alloc: true, pod: 0xa0249adc0c01cf50, env: 0x88be0fba4063a209}},
				}))
			})
			t.Run("3 deactivate", func(t *testing.T) {
				arbiter1.ConsumeMessage(bus.NewMessage("", nil))
				fixture.WaitNoErrorT10(t, evaluator1.assertFn(map[string][]dummyEvRecord{
					"first":  {{alloc: true, pod: 0x1c2120eda6a20a7f, env: 0x88be0fba4063a209}},
					"second": {{alloc: true, pod: 0xadbac957c6a4641f, env: 0x88be0fba4063a209}},
					"third":  {{alloc: false, pod: 0x0, env: 0x0}, {alloc: true, pod: 0xa0249adc0c01cf50, env: 0x88be0fba4063a209}},
				}))
			})
			t.Run("4 remove third", func(t *testing.T) {
				var buffers lib.StaticBuffers
				var registry manifest.PodSlice
				assert.NoError(t, buffers.ReadFiles("testdata/sink_test_ConsumeRegistry_4.hcl"))
				assert.NoError(t, registry.Unmarshal(manifest.PrivateNamespace, buffers.GetReaders()...))
				sink.ConsumeRegistry(registry)
				fixture.WaitNoErrorT10(t, evaluator1.assertFn(map[string][]dummyEvRecord{
					"first":  {{alloc: true, pod: 0x1c2120eda6a20a7f, env: 0x88be0fba4063a209}},
					"second": {{alloc: true, pod: 0xadbac957c6a4641f, env: 0x88be0fba4063a209}},
					"third":  {{alloc: false, pod: 0x0, env: 0x0}, {alloc: true, pod: 0xa0249adc0c01cf50, env: 0x88be0fba4063a209}, {alloc: false, pod: 0x0, env: 0x0}},
				}))
			})
			t.Run("5 activate", func(t *testing.T) {
				arbiter1.ConsumeMessage(bus.NewMessage("", map[string]string{
					"meta.first":  "true",
					"meta.second": "true",
				}))
				fixture.WaitNoErrorT10(t, evaluator1.assertFn(map[string][]dummyEvRecord{
					"first":  {{alloc: true, pod: 0x1c2120eda6a20a7f, env: 0x88be0fba4063a209}, {alloc: true, pod: 0x1c2120eda6a20a7f, env: 0x88be0fba4063a209}},
					"second": {{alloc: true, pod: 0xadbac957c6a4641f, env: 0x88be0fba4063a209}, {alloc: true, pod: 0xadbac957c6a4641f, env: 0x88be0fba4063a209}},
					"third":  {{alloc: false, pod: 0x0, env: 0x0}, {alloc: true, pod: 0xa0249adc0c01cf50, env: 0x88be0fba4063a209}, {alloc: false, pod: 0x0, env: 0x0}},
				}))

			})
			t.Run("6 drain", func(t *testing.T) {
				arbiter1.ConsumeMessage(bus.NewMessage("", map[string]string{
					"drain": "true",
				}))
				fixture.WaitNoErrorT10(t, evaluator1.assertFn(map[string][]dummyEvRecord{
					"first":  {{alloc: true, pod: 0x1c2120eda6a20a7f, env: 0x88be0fba4063a209}, {alloc: true, pod: 0x1c2120eda6a20a7f, env: 0x88be0fba4063a209}, {alloc: false, pod: 0x0, env: 0x0}},
					"second": {{alloc: true, pod: 0xadbac957c6a4641f, env: 0x88be0fba4063a209}, {alloc: true, pod: 0xadbac957c6a4641f, env: 0x88be0fba4063a209}, {alloc: false, pod: 0x0, env: 0x0}},
					"third":  {{alloc: false, pod: 0x0, env: 0x0}, {alloc: true, pod: 0xa0249adc0c01cf50, env: 0x88be0fba4063a209}, {alloc: false, pod: 0x0, env: 0x0}},
				}))
			})
			t.Run("7 remove drain", func(t *testing.T) {
				arbiter1.ConsumeMessage(bus.NewMessage("", map[string]string{
					"meta.first":  "true",
					"meta.second": "true",
				}))
				fixture.WaitNoErrorT10(t, evaluator1.assertFn(map[string][]dummyEvRecord{
					"first":  {{alloc: true, pod: 0x1c2120eda6a20a7f, env: 0x88be0fba4063a209}, {alloc: true, pod: 0x1c2120eda6a20a7f, env: 0x88be0fba4063a209}, {alloc: false, pod: 0x0, env: 0x0}, {alloc: true, pod: 0x1c2120eda6a20a7f, env: 0x88be0fba4063a209}},
					"second": {{alloc: true, pod: 0xadbac957c6a4641f, env: 0x88be0fba4063a209}, {alloc: true, pod: 0xadbac957c6a4641f, env: 0x88be0fba4063a209}, {alloc: false, pod: 0x0, env: 0x0}, {alloc: true, pod: 0xadbac957c6a4641f, env: 0x88be0fba4063a209}},
					"third":  {{alloc: false, pod: 0x0, env: 0x0}, {alloc: true, pod: 0xa0249adc0c01cf50, env: 0x88be0fba4063a209}, {alloc: false, pod: 0x0, env: 0x0}},
				}))
			})

			sv.Close()
			sv.Wait()
		})
	}

}
