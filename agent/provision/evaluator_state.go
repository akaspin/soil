package provision

import (
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"sync"
)

type EvaluatorState struct {
	log *logx.Log
	mu  sync.Mutex

	finished   map[string]*allocation.Pod // Finished evaluations
	inProgress map[string]*allocation.Pod // Evaluations in progress
	pending    map[string]*allocation.Pod // Pending allocations
}

func NewEvaluatorState(log *logx.Log, recovered allocation.Recovery) (s *EvaluatorState) {
	s = &EvaluatorState{
		log:        log.WithTags("evaluator", "state"),
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
	s.log.Tracef(`submit: %s`, name)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pending[name] = pod
	s.log.Tracef(`submit: registered pending %s`, name)
	next = s.next()
	return
}

// Commit in progress evaluation
func (s *EvaluatorState) Commit(name string) (next []*Evaluation) {
	s.log.Tracef(`commit: %s`, name)
	s.mu.Lock()
	defer s.mu.Unlock()

	if in := s.inProgress[name]; in != nil {
		s.finished[name] = in
		s.log.Tracef(`%s promoted to finished`, name)
	} else {
		delete(s.finished, name)
		s.log.Tracef(`%s removed from finished`, name)
	}
	delete(s.inProgress, name)
	s.log.Tracef(`%s removed from in progress`, name)
	next = s.next()
	return
}

func (s *EvaluatorState) next() (next []*Evaluation) {
LOOP:
	for pendingName, pending := range s.pending {
		if inProgress, exists := s.inProgress[pendingName]; exists {
			// blocked by inProgress
			if allocation.IsEqual(inProgress, pending) {
				delete(s.pending, pendingName)
				s.log.Tracef(`pending %s removed: equal to in progress`, pendingName)
			}
			s.log.Tracef(`skip promote pending %s: in progress`, pendingName)
			continue LOOP
		}
		// check for blockers
		for _, finished := range s.finished {
			if finished.Name != pendingName {
				if err := allocation.IsBlocked(finished, pending); err != nil {
					s.log.Warning(err)
					continue LOOP
				}
			}
		}
		// not blocked
		finished := s.finished[pendingName]
		if allocation.IsEqual(finished, pending) {
			delete(s.pending, pendingName)
			s.log.Tracef(`pending %s removed: equal to finished`, pendingName)
			continue LOOP
		}
		s.inProgress[pendingName] = pending
		delete(s.pending, pendingName)
		next = append(next, NewEvaluation(s.finished[pendingName], pending))
		s.log.Tracef(`pending %s promoted to in progress`, pendingName)
	}
	s.log.Debugf(`next: %s`, next)
	return
}
