// +build ide test_unit

package resource2_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/resource2"
	"github.com/akaspin/soil/fixture"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEvaluator_Configure(t *testing.T) {
	t.Skip()
	ctx := context.Background()
	log := logx.GetLog("test")

	runTest := func(t *testing.T, config []resource2.Config, state allocation.Recovery, downstream, upstream []bus.Message) {
		t.Helper()
		ctx1, cancel := context.WithCancel(ctx)
		defer cancel()

		downstreamCons := bus.NewTestingConsumer(ctx1)
		upstreamCons := bus.NewTestingConsumer(ctx1)
		evaluator := resource2.NewEvaluator(ctx, log, resource2.EvaluatorConfig{}, state, downstreamCons, upstreamCons)
		assert.NoError(t, evaluator.Open())
		evaluator.Configure(config)

		fixture.WaitNoError(t, fixture.DefaultWaitConfig(), downstreamCons.ExpectMessagesFn(downstream...))
		fixture.WaitNoError(t, fixture.DefaultWaitConfig(), upstreamCons.ExpectMessagesFn(upstream...))

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
			[]resource2.Config{
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
			[]resource2.Config{
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
			[]resource2.Config{
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
