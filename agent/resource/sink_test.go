// build ide test_unit

package resource_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/resource"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	"github.com/akaspin/soil/lib"
	"github.com/akaspin/soil/manifest"
)

func TestSink_Flow_NoRecovery(t *testing.T) {
	ctx := context.Background()
	log := logx.GetLog("test")
	waitTime := time.Millisecond * 300
	arbiter := scheduler.NewArbiter(ctx, log, "resource", scheduler.ArbiterConfig{})
	arbiterCompositePipe := bus.NewCompositePipe("private", arbiter, "resource")

	downstreamCons := &bus.DummyConsumer{}
	checkCons := &bus.DummyConsumer{}
	upstream := bus.NewSimplePipe(nil, arbiterCompositePipe, checkCons)

	evaluator := resource.NewEvaluator(ctx, log, resource.EvaluatorConfig{}, nil, downstreamCons, upstream)
	sink := scheduler.NewSink(ctx, log, nil, scheduler.NewBoundedEvaluator(
		arbiter, evaluator,
	))
	sv := supervisor.NewChain(ctx,
		supervisor.NewChain(ctx,
			supervisor.NewGroup(ctx, evaluator, arbiter),
			sink,
		),
	)
	assert.NoError(t, sv.Open())

	evaluator.Configure()
	time.Sleep(waitTime)

	checkCons.AssertMessages(t, bus.NewMessage("resource", map[string]string{}))
	downstreamCons.AssertMessages(t, bus.NewMessage("resource", map[string]string{}))

	sv.Close()
	sv.Wait()
}

func TestSink_Flow(t *testing.T) {
	ctx := context.Background()
	log := logx.GetLog("test")
	waitTime := time.Millisecond * 300
	arbiter := scheduler.NewArbiter(ctx, log, "resource", scheduler.ArbiterConfig{})
	arbiterCompositePipe := bus.NewCompositePipe("private", arbiter, "resource")

	downstreamCons := &bus.DummyConsumer{}
	checkCons := &bus.DummyConsumer{}
	upstream := bus.NewSimplePipe(nil, arbiterCompositePipe, checkCons)

	var state allocation.Recovery
	assert.NoError(t, state.FromFilesystem(
		allocation.SystemPaths{
			Local:   "testdata/etc",
			Runtime: "testdata/TestEvaluator_Configure",
		},
		allocation.GetZeroDiscoveryFunc(
			"testdata/TestSink_Flow/pod-test-1.service",
			"testdata/TestSink_Flow/pod-test-2.service",
		)))
	evaluator := resource.NewEvaluator(ctx, log, resource.EvaluatorConfig{}, state, downstreamCons, upstream)
	sink := scheduler.NewSink(ctx, log, state, scheduler.NewBoundedEvaluator(
		arbiter, evaluator,
	))
	sv := supervisor.NewChain(ctx,
		supervisor.NewChain(ctx,
			supervisor.NewGroup(ctx, evaluator, arbiter),
			sink,
		),
	)
	assert.NoError(t, sv.Open())

	t.Run("0 configure with recovery", func(t *testing.T) {
		evaluator.Configure([]resource.Config{
			{
				Nature: "dummy",
				Kind:   "fake1",
			},
			{
				Nature: "dummy",
				Kind:   "fake2",
			},
		}...)
		time.Sleep(waitTime)

		checkCons.AssertMessages(t,
			bus.NewMessage("resource", map[string]string{
				"request.fake1.allow": "true",
				"request.fake2.allow": "true",
			}),
		)
		downstreamCons.AssertMessages(t,
			bus.NewMessage("resource", map[string]string{
				"fake1.test-1.1.allocated": "true",
				"fake1.test-1.1.fixed":     "8080",
				"fake1.test-1.1.__values":  "{\"allocated\":\"true\",\"fixed\":\"8080\"}",
				"fake1.test-1.2.allocated": "true",
				"fake1.test-1.2.__values":  "{\"allocated\":\"true\"}",
				"fake2.test-1.1.allocated": "true",
				"fake2.test-1.1.__values":  "{\"allocated\":\"true\"}",
			}),
		)
	})
	t.Run("1 submit", func(t *testing.T) {
		var buffers lib.StaticBuffers
		var registry manifest.Registry
		assert.NoError(t, buffers.ReadFiles("testdata/TestSink_Flow/1.hcl"))
		assert.NoError(t, registry.Unmarshal("private", buffers.GetReaders()...))
		sink.ConsumeRegistry(registry)
		time.Sleep(waitTime)
		downstreamCons.AssertLastMessage(t,
			bus.NewMessage("resource", map[string]string{
				"fake1.test-1.1.allocated": "true",
				"fake1.test-1.1.__values":  "{\"allocated\":\"true\"}",
			}),
		)
	})

	sv.Close()
	sv.Wait()
}
