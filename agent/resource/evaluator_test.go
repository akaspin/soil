// +build ide test_unit

package resource_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/resource"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewEvaluator(t *testing.T) {
	// recover state

	paths := allocation.SystemPaths{
		Local:   "testdata/etc",
		Runtime: "testdata/TestEvaluator_Configure",
	}
	var state allocation.Recovery
	assert.NoError(t, state.FromFilesystem(paths, allocation.GetZeroDiscoveryFunc(
		"testdata/TestEvaluator_Configure/pod-test-1.service",
		"testdata/TestEvaluator_Configure/pod-test-2.service",
	)))

	ctx := context.Background()
	log := logx.GetLog("test")
	cons := &bus.DummyConsumer{}

	evaluator := resource.NewEvaluator(ctx, log, resource.EvaluatorConfig{}, state, cons)

	assert.NoError(t, evaluator.Open())

	config := []resource.Config{
		{
			Kind:   "fake1",
			Nature: "test",
		},
		{
			Kind:   "fake2",
			Nature: "test",
		},
	}
	evaluator.Configure(config...)
	time.Sleep(time.Millisecond * 100)

	//pretty.Log(evaluator)

	evaluator.Close()
	evaluator.Wait()
}
