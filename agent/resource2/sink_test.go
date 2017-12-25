// +build ide test_unit

package resource2_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/resource2"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/fixture"
	"github.com/akaspin/soil/lib"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSink_Flow_NoRecovery(t *testing.T) {
	t.Skip()
	ctx := context.Background()
	log := logx.GetLog("test")
	arbiter := scheduler.NewArbiter(ctx, log, "resource", scheduler.ArbiterConfig{})
	arbiterCompositePipe := bus.NewCompositePipe("private", log, arbiter, "resource")

	consCtx, consCancel := context.WithCancel(context.Background())
	defer consCancel()
	downstreamCons := bus.NewTestingConsumer(consCtx)
	checkCons := bus.NewTestingConsumer(consCtx)

	upstream := bus.NewTeePipe(arbiterCompositePipe, checkCons)

	evaluator := resource2.NewEvaluator(ctx, log, resource2.EvaluatorConfig{}, nil, downstreamCons, upstream)
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

	evaluator.Configure(nil)

	fixture.WaitNoError(t, fixture.DefaultWaitConfig(), checkCons.ExpectMessagesFn(
		bus.NewMessage("resource", map[string]string{}),
	))
	fixture.WaitNoError(t, fixture.DefaultWaitConfig(), downstreamCons.ExpectMessagesFn(
		bus.NewMessage("resource", map[string]string{}),
	))

	sv.Close()
	sv.Wait()
}

func TestSink_Flow(t *testing.T) {
	t.Skip()
	ctx := context.Background()
	log := logx.GetLog("test")
	waitTime := time.Millisecond * 300
	arbiter := scheduler.NewArbiter(ctx, log, "resource", scheduler.ArbiterConfig{})
	arbiterCompositePipe := bus.NewCompositePipe("private", log, arbiter, "resource")

	consCtx, consCancel := context.WithCancel(context.Background())
	defer consCancel()
	downstreamCons := bus.NewTestingConsumer(consCtx)
	checkCons := bus.NewTestingConsumer(consCtx)
	upstream := bus.NewTeePipe(arbiterCompositePipe, checkCons)

	var state allocation.PodSlice
	assert.NoError(t, state.FromFilesystem(
		allocation.SystemPaths{
			Local:   "testdata/etc",
			Runtime: "testdata/TestEvaluator_Configure",
		},
		allocation.GetZeroDiscoveryFunc(
			"testdata/TestSink_Flow/pod-test-1.service",
			"testdata/TestSink_Flow/pod-test-2.service",
		)))
	evaluator := resource2.NewEvaluator(ctx, log, resource2.EvaluatorConfig{}, state, downstreamCons, upstream)
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
		evaluator.Configure([]resource2.Config{
			{
				Nature: "dummy",
				Kind:   "fake1",
			},
			{
				Nature: "dummy",
				Kind:   "fake2",
			},
		})

		fixture.WaitNoError(t, fixture.DefaultWaitConfig(), checkCons.ExpectMessagesFn(
			bus.NewMessage("resource", map[string]string{
				"request.fake1.allow": "true",
				"request.fake2.allow": "true",
			}),
		))
		fixture.WaitNoError(t, fixture.DefaultWaitConfig(), downstreamCons.ExpectMessagesFn(
			bus.NewMessage("resource", map[string]string{
				"fake1.test-1.1.allocated": "true",
				"fake1.test-1.1.fixed":     "8080",
				"fake1.test-1.1.__values":  "{\"allocated\":\"true\",\"fixed\":\"8080\"}",
				"fake1.test-1.2.allocated": "true",
				"fake1.test-1.2.__values":  "{\"allocated\":\"true\"}",
				"fake2.test-1.1.allocated": "true",
				"fake2.test-1.1.__values":  "{\"allocated\":\"true\"}",
			}),
		))
	})
	t.Run("1 submit", func(t *testing.T) {
		var buffers lib.StaticBuffers
		var registry manifest.Pods
		assert.NoError(t, buffers.ReadFiles("testdata/TestSink_Flow/1.hcl"))
		assert.NoError(t, registry.Unmarshal("private", buffers.GetReaders()...))
		sink.ConsumeRegistry(registry)
		time.Sleep(waitTime)
		fixture.WaitNoError(t, fixture.DefaultWaitConfig(), downstreamCons.ExpectLastMessageFn(
			bus.NewMessage("resource", map[string]string{
				"fake1.test-1.1.allocated": "true",
				"fake1.test-1.1.__values":  "{\"allocated\":\"true\"}",
			}),
		))
	})

	sv.Close()
	sv.Wait()
}
