// +build ide test_unit

package scheduler_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"sort"
	"sync"
	"testing"
	"time"
)

type testStateHolder struct{}

func (*testStateHolder) GetState() allocation.State {
	return nil
}

type testEvRecord struct {
	name string
	pod  *manifest.Pod
	env  map[string]string
}

type testSimpleEvaluator struct {
	name    string
	mu      sync.Mutex
	records map[string]testEvRecord
	names   []string
}

func newTestSimpleEvaluator(name string) (e *testSimpleEvaluator) {
	e = &testSimpleEvaluator{
		name:    name,
		records: map[string]testEvRecord{},
	}
	return
}

func (e *testSimpleEvaluator) GetConstraint(pod *manifest.Pod) (res manifest.Constraint) {
	res = pod.GetResourceAllocationConstraint()
	return
}

func (e *testSimpleEvaluator) Allocate(name string, pod *manifest.Pod, env map[string]string) {
	e.submit(name, pod, env)
}

func (e *testSimpleEvaluator) Deallocate(name string) {
	e.submit(name, nil, nil)
}

func (e *testSimpleEvaluator) submit(name string, pod *manifest.Pod, env map[string]string) {
	go func() {
		e.mu.Lock()
		defer e.mu.Unlock()
		record := testEvRecord{name, pod, env}
		e.records[name] = record
		e.names = []string{}
		for n, rec := range e.records {
			if rec.pod != nil {
				e.names = append(e.names, n)
			}
		}
		sort.Strings(e.names)
	}()
}

//
type testDownstreamEvaluator struct {
}

func TestSink_TwoManagedEvaluators(t *testing.T) {
	ctx := context.Background()
	log := logx.GetLog("test")

	evaluator1 := newTestSimpleEvaluator("1")
	evaluator2 := newTestSimpleEvaluator("2")

	sources := []scheduler.ManagerSource{
		scheduler.NewManagerSource("meta", false, nil, "private", "public"),
		scheduler.NewManagerSource("resource", false, nil, "private", "public"),
	}

	manager1 := scheduler.NewManager(ctx, log, "1", sources...)
	manager2 := scheduler.NewManager(ctx, log, "2", sources...)
	sink := scheduler.NewSink(ctx, log, &testStateHolder{},
		scheduler.NewManagedEvaluator(manager1, evaluator1),
		scheduler.NewManagedEvaluator(manager2, evaluator2),
	)
	metaProducer := bus.NewFlatMap(true, "meta", manager1, manager2)
	resourceProducer := bus.NewFlatMap(true, "resource", manager1, manager2)

	sv := supervisor.NewChain(ctx,
		manager1, manager2, sink,
	)
	assert.NoError(t, sv.Open())

	var registry manifest.Registry
	err := registry.UnmarshalFiles("private", "testdata/sink_test_0.hcl")
	assert.NoError(t, err)
	sink.ConsumeRegistry("private", registry)

	t.Run("0 activate meta", func(t *testing.T) {
		metaProducer.Set(map[string]string{
			"first":  "true",
			"second": "true",
		})
		time.Sleep(time.Millisecond * 100)
		assert.Equal(t, evaluator1.names, []string(nil))
		assert.Equal(t, evaluator2.names, []string(nil))
	})
	t.Run("1 activate resource", func(t *testing.T) {
		resourceProducer.Set(map[string]string{})
		time.Sleep(time.Millisecond * 100)

		assert.Equal(t, evaluator1.names, []string{"first", "second"})
		assert.Equal(t, evaluator2.names, []string{"first", "second"})
	})
	t.Run("2 allocate resource", func(t *testing.T) {
		resourceProducer.Set(map[string]string{
			"port.third.8080.allocated": "true",
		})
		time.Sleep(time.Millisecond * 200)
		assert.Equal(t, evaluator1.names, []string{"first", "second", "third"})
		assert.Equal(t, evaluator2.names, []string{"first", "second", "third"})
	})
	t.Run("3 deallocate resource", func(t *testing.T) {
		resourceProducer.Set(map[string]string{})
		time.Sleep(time.Millisecond * 100)
		assert.Equal(t, evaluator1.names, []string{"first", "second"})
		assert.Equal(t, evaluator2.names, []string{"first", "second"})
	})
	t.Run("4 divert", func(t *testing.T) {
		manager2.ConsumeMessage(bus.NewMessage("resource", map[string]string{
			"port.third.8080.allocated": "true",
		}))
		time.Sleep(time.Millisecond * 100)
		assert.Equal(t, evaluator1.names, []string{"first", "second"})
		assert.Equal(t, evaluator2.names, []string{"first", "second", "third"})
	})

	sv.Close()
	sv.Wait()
}

// Test Sink with
func TestSink_Stacked(t *testing.T) {
	//ctx := context.Background()
	//log := logx.GetLog("test")

}
