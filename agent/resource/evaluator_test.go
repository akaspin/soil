// build ide test_unit

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

func TestEvaluator_Configure(t *testing.T) {
	ctx := context.Background()
	log := logx.GetLog("test")

	runTest := func(t *testing.T, config []resource.Config, state allocation.Recovery, downstream, upstream []bus.Message) {
		t.Helper()
		downstreamCons := &bus.DummyConsumer{}
		upstreamCons := &bus.DummyConsumer{}
		evaluator := resource.NewEvaluator(ctx, log, resource.EvaluatorConfig{}, state, downstreamCons, upstreamCons)
		assert.NoError(t, evaluator.Open())
		evaluator.Configure(config...)
		time.Sleep(time.Millisecond * 200)

		downstreamCons.AssertMessages(t, downstream...)
		upstreamCons.AssertMessages(t, upstream...)

		evaluator.Close()
		evaluator.Wait()
	}

	t.Run("0 empty with no configs", func(t *testing.T) {
		runTest(t, nil, nil,
			[]bus.Message{bus.NewMessage("resource", map[string]string{})},
			[]bus.Message{bus.NewMessage("resource", map[string]string{})})
	})
	t.Run("1 empty with configs", func(t *testing.T) {
		runTest(t,
			[]resource.Config{
				{
					Kind:   "fake1",
					Nature: "dummy",
				},
				{
					Kind:   "fake2",
					Nature: "dummy",
				},
			},
			nil,
			[]bus.Message{
				bus.NewMessage("resource", map[string]string{}),
			},
			[]bus.Message{bus.NewMessage("resource",
				map[string]string{
					"request.fake1.allow": "true",
					"request.fake2.allow": "true",
				})},
		)
	})
	var state allocation.Recovery
	assert.NoError(t, state.FromFilesystem(
		allocation.SystemPaths{
			Local:   "testdata/etc",
			Runtime: "testdata/TestEvaluator_Configure",
		},
		allocation.GetZeroDiscoveryFunc(
			"testdata/TestEvaluator_Configure/pod-test-1.service",
			"testdata/TestEvaluator_Configure/pod-test-2.service",
		)))

	t.Run("0 configs and allocations", func(t *testing.T) {
		runTest(t,
			[]resource.Config{
				{
					Kind:   "fake1",
					Nature: "dummy",
				},
				{
					Kind:   "fake2",
					Nature: "dummy",
				},
			},
			state,
			[]bus.Message{
				bus.NewMessage("resource", map[string]string{
					"fake1.test-1.1.allocated": "true",
					"fake1.test-1.1.fixed":     "8080",
					"fake1.test-1.1.__values":  "{\"allocated\":\"true\",\"fixed\":\"8080\"}",
					"fake1.test-1.2.allocated": "true",
					"fake1.test-1.2.__values":  "{\"allocated\":\"true\"}",
					"fake2.test-1.1.allocated": "true",
					"fake2.test-1.1.__values":  "{\"allocated\":\"true\"}",
				}),
			},
			[]bus.Message{bus.NewMessage("resource",
				map[string]string{
					"request.fake1.allow": "true",
					"request.fake2.allow": "true",
				})},
		)
	})
	t.Run("0 configs and extra allocations", func(t *testing.T) {
		runTest(t,
			[]resource.Config{
				{
					Kind:   "fake1",
					Nature: "dummy",
				},
			},
			state,
			[]bus.Message{
				bus.NewMessage("resource", map[string]string{
					"fake1.test-1.1.allocated": "true",
					"fake1.test-1.1.fixed":     "8080",
					"fake1.test-1.1.__values":  "{\"allocated\":\"true\",\"fixed\":\"8080\"}",
					"fake1.test-1.2.allocated": "true",
					"fake1.test-1.2.__values":  "{\"allocated\":\"true\"}",
				}),
			},
			[]bus.Message{bus.NewMessage("resource",
				map[string]string{
					"request.fake1.allow": "true",
				})},
		)
	})

}
