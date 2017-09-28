// +build ide test_unit

package scheduler_test

import (
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func makeAllocations(t *testing.T, path string) (recovered []*allocation.Pod) {
	t.Helper()
	var pods manifest.Registry
	err := pods.UnmarshalFiles("private", path)
	assert.NoError(t, err)
	for _, pod := range pods {
		alloc, _ := allocation.NewFromManifest(pod, allocation.DefaultSystemDPaths(), map[string]string{})
		recovered = append(recovered, alloc)
	}
	return
}

func zeroEvaluatorState(t *testing.T) (s *scheduler.EvaluatorState) {
	t.Helper()
	recovered := makeAllocations(t, "testdata/evaluator_state_test_0.hcl")
	s = scheduler.NewEvaluatorState(recovered)
	return
}

func TestEvaluatorState(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		state := zeroEvaluatorState(t)
		// simple submit
		next := state.Submit("pod-1", makeAllocations(t, "testdata/evaluator_state_test_1.hcl")[0])
		assert.Len(t, next, 0)
	})
	t.Run("2", func(t *testing.T) {
		state := zeroEvaluatorState(t)
		// submit blocking pod
		next := state.Submit("pod-3", makeAllocations(t, "testdata/evaluator_state_test_2.hcl")[0])
		assert.Len(t, next, 0)
		// remove pod-1 (unblock pod-3)
		next = state.Submit("pod-1", nil)
		assert.Len(t, next, 1)
		assert.NotNil(t, next[0].Left)
		assert.Equal(t, next[0].Left.Name, "pod-1")
		assert.Nil(t, next[0].Right)
		next = state.Commit("pod-1")
		assert.Len(t, next, 1)
		assert.Nil(t, next[0].Left)
		assert.NotNil(t, next[0].Right)
		assert.Equal(t, next[0].Right.Name, "pod-3")
	})
	t.Run("3", func(t *testing.T) {
		state := zeroEvaluatorState(t)
		// simple submit
		next := state.Submit("pod-3", nil)
		assert.Len(t, next, 0)
	})
}
