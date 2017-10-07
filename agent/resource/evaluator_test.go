package resource_test

import (
	"testing"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/stretchr/testify/assert"
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/resource"
	"github.com/kr/pretty"
)

func TestEvaluator_Configure(t *testing.T) {
	paths := allocation.SystemPaths{
		Local:   "testdata/etc",
		Runtime: "testdata/TestEvaluator_Configure",
	}
	var state allocation.State
	assert.NoError(t, state.Discover(paths, allocation.GetZeroDiscoveryFunc(
		"testdata/TestEvaluator_Configure/pod-test-1.service",
		"testdata/TestEvaluator_Configure/pod-test-2.service",
	)))

	ctx := context.Background()
	log := logx.GetLog("test")
	cons := &bus.TestConsumer{}

	evaluator := resource.NewEvaluator(ctx, log, state, cons)

	assert.NoError(t, evaluator.Open())
	evaluator.Configure([]resource.Config{
		{
			Type: "fake",
			Name: "fake1",
			Properties: map[string]interface{}{},
		},
	}...)

	pretty.Log(cons)

	evaluator.Close()
	evaluator.Wait()
}
