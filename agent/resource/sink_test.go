package resource_test

import (
	"testing"
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/stretchr/testify/assert"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/resource"
	"github.com/akaspin/supervisor"
	"time"
	"github.com/kr/pretty"
)

func TestSink_Flow(t *testing.T) {
	t.Skip()
	ctx := context.Background()
	log := logx.GetLog("test")
	waitTime := time.Millisecond * 300
	arbiter := scheduler.NewArbiter(ctx, log, "resource", nil)
	arbiterCompositePipe := bus.NewCompositePipe("0", arbiter, "resource")

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

	t.Run("0 configure", func(t *testing.T) {
		evaluator.Configure([]resource.Config{
			{
				Nature: "dummy",
				Kind: "fake1",
			},
			{
				Nature: "dummy",
				Kind: "fake2",
			},
		}...)
		time.Sleep(waitTime)

		pretty.Log(checkCons)
		pretty.Log(downstreamCons)
	})

	sv.Close()
	sv.Wait()
}
