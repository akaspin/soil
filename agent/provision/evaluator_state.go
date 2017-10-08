package provision

import (
	"github.com/akaspin/soil/agent/allocation"
	"sync"
)

type EvaluatorState struct {
	mu sync.RWMutex

	// Finished evaluations
	finished map[string]*allocation.Pod

	// Evaluations in progress
	inProgress map[string]*allocation.Pod

	// Pending allocations
	pending map[string]*allocation.Pod
}

func NewEvaluatorState(recovered allocation.Recovery) (s *EvaluatorState) {
	s = &EvaluatorState{
		finished:   map[string]*allocation.Pod{},
		inProgress: map[string]*allocation.Pod{},
		pending:    map[string]*allocation.Pod{},
	}
	for _, pod := range recovered {
		s.finished[pod.Name] = pod
	}
	return
}

// Submit allocation to state. Returns allocations ready to execute.
func (s *EvaluatorState) Submit(name string, pod *allocation.Pod) (next []*Evaluation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pending[name] = pod
	next = s.next()
	return
}

// Commit in progress evaluation
func (s *EvaluatorState) Commit(name string) (next []*Evaluation) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if in := s.inProgress[name]; in != nil {
		s.finished[name] = in
	} else {
		delete(s.finished, name)
	}
	delete(s.inProgress, name)
	next = s.next()
	return
}

func (s *EvaluatorState) GetState() (state allocation.Recovery) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	mapped := map[string]*allocation.Pod{}

	for _, what := range []map[string]*allocation.Pod{
		s.pending, s.inProgress, s.finished,
	} {
		for k, v := range what {
			if _, ok := mapped[k]; !ok {
				if v == nil {
					mapped[k] = nil
					continue
				}
				mapped[k] = v
			}
		}
	}
	for _, v := range mapped {
		if v != nil {
			state = append(state, v)
		}
	}
	return
}

func (s *EvaluatorState) next() (next []*Evaluation) {
LOOP:
	for pendingName, pending := range s.pending {
		if inProgress, exists := s.inProgress[pendingName]; exists {
			// blocked by inProgress
			if allocation.IsEqual(inProgress, pending) {
				delete(s.pending, pendingName)
			}
			continue LOOP
		}
		// check for blockers
		for _, finished := range s.finished {
			if finished.Name != pendingName && allocation.IsBlocked(finished, pending) {
				continue LOOP
			}
		}
		// not blocked
		finished := s.finished[pendingName]
		if allocation.IsEqual(finished, pending) {
			delete(s.pending, pendingName)
			continue LOOP
		}
		s.inProgress[pendingName] = pending
		delete(s.pending, pendingName)
		next = append(next, NewEvaluation(s.finished[pendingName], pending))
	}
	return
}
