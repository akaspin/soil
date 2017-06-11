package scheduler

import "github.com/akaspin/soil/agent/allocation"

type EvaluatorState struct {

}


// Submit allocation to state. Returns allocations ready to execute.
func (s *EvaluatorState) Submit(name string, pod *allocation.Pod) (next map[string]*allocation.Pod) {
	return
}

// Commit in progress evaluation
func (s *EvaluatorState) Commit(name string) (next map[string]*allocation.Pod) {
	return
}