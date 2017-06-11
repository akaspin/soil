package scheduler_test

import (
	"testing"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/stretchr/testify/assert"
)

func makeAllocations(path string) (recovered []*allocation.Pod) {
	pods, _ := manifest.ParseFromFiles("private", path)
	for _, pod := range pods {
		alloc, _ := allocation.NewFromManifest(pod, map[string]string{}, 0)
		recovered = append(recovered, alloc)
	}
	return
}

func zeroEvaluatorState() (s *scheduler.EvaluatorState) {
	recovered := makeAllocations("testdata/evaluator_state_test_0.hcl")
	s = scheduler.NewEvaluatorState(recovered)
	return
}

func TestEvaluatorState(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		state := zeroEvaluatorState()
		// simple submit
		next := state.Submit("pod-1", makeAllocations("testdata/evaluator_state_test_1.hcl")[0])
		assert.Len(t, next, 0)
	})
	t.Run("2", func(t *testing.T) {
		state := zeroEvaluatorState()
		// submit blocking pod
		next := state.Submit("pod-3", makeAllocations("testdata/evaluator_state_test_2.hcl")[0])
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
		state := zeroEvaluatorState()
		// simple submit
		next := state.Submit("pod-3", nil)
		assert.Len(t, next, 0)
	})
}
