// +build ide test_unit

package provision_test

import (
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/provision"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEvaluatorState_Submit(t *testing.T) {
	zeroEvaluatorState := func(t *testing.T) (s *provision.EvaluatorState) {
		t.Helper()
		recovered := makeAllocations(t, "testdata/evaluator_state_test_0.hcl")
		s = provision.NewEvaluatorState(logx.GetLog("test"), recovered)
		return
	}

	t.Run("1 submit pod-1", func(t *testing.T) {
		state := zeroEvaluatorState(t)
		// simple submit
		next := state.Submit("pod-1", makeAllocations(t, "testdata/evaluator_state_test_1.hcl")[0])
		assert.Len(t, next, 0, "pod-1 should be not changed")
	})
	t.Run("2 submit blocking pod-3 with unit pod-1-1)", func(t *testing.T) {
		state := zeroEvaluatorState(t)
		next := state.Submit("pod-3", makeAllocations(t, "testdata/evaluator_state_test_2.hcl")[0])

		assert.Len(t, next, 0, "pod-3 should be blocked")

		next = state.Submit("pod-1", nil)
		assert.Len(t, next, 1, "pod-1 should be destroyed")
		assert.NotNil(t, next[0].Left)
		assert.Equal(t, next[0].Left.Name, "pod-1")
		assert.Nil(t, next[0].Right)

		next = state.Commit("pod-1")
		assert.Len(t, next, 1)
		assert.Nil(t, next[0].Left)
		assert.NotNil(t, next[0].Right)
		assert.Equal(t, next[0].Right.Name, "pod-3")
	})
	t.Run("3 submit non-existent pod", func(t *testing.T) {
		state := zeroEvaluatorState(t)
		next := state.Submit("pod-3", nil)
		assert.Len(t, next, 0)
	})
	t.Run("4 submit changed pod-1", func(t *testing.T) {
		state := zeroEvaluatorState(t)
		next := state.Submit("pod-1", makeAllocations(t, "testdata/evaluator_state_test_4.hcl")[0])
		assert.Len(t, next, 1, "pod-1 should be evaluated")
	})
}
