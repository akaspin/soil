package executor

import (
	"github.com/akaspin/soil/agent/scheduler/allocation"
	"github.com/pkg/errors"
	"sync"
)

var (
	AllocationNotFoundError = errors.New("allocation is not found")
	AllocationNotUnique     = errors.New("allocation is not unique")
)

type State struct {
	ready   map[string]*allocation.Pod
	active  map[string]*allocation.Pod
	pending map[string]*allocation.Pod

	mu      *sync.Mutex
}

func NewState(initial []*allocation.Pod) (s *State) {
	s = &State{
		ready:   map[string]*allocation.Pod{},
		active:  map[string]*allocation.Pod{},
		pending: map[string]*allocation.Pod{},
		mu:      &sync.Mutex{},
	}
	for _, a := range initial {
		s.ready[a.PodHeader.Name] = a
	}
	return
}

// Submit allocation to pending. Use <nil> for destroy.
// Submit returns ok if state actually submitted.
func (s *State) Submit(name string, pending *allocation.Pod) (ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	latest := s.getLatest(name)
	if !allocation.ComparePods(latest, pending) {
		s.pending[name] = pending
		ok = true
	}
	return
}

// Promote allocation from pending to active and return ready and active pair.
// or error if evaluation is not possible at this time.
func (s *State) Promote(name string) (ready, active *allocation.Pod, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var ok bool
	if _, ok = s.active[name]; ok {
		err = errors.Wrapf(AllocationNotUnique, "can't promote pending %s", name)
		return
	}
	if active, ok = s.pending[name]; !ok {
		err = errors.Wrapf(AllocationNotFoundError, "can't promote pending %s", name)
		return
	}

	ready = s.ready[name]
	s.active[name] = active
	delete(s.pending, name)

	return
}

// Commit active to ready
func (s *State) Commit(name string, failures []error) (destroyed bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	active, ok := s.active[name]
	if !ok {
		err = errors.Wrapf(AllocationNotFoundError, "can't commit %s", name)
		return
	}
	destroyed = active == nil
	if destroyed {
		delete(s.ready, name)
	} else {
		s.ready[name] = active
	}
	delete(s.active, name)
	return
}

// list
func (s *State) ListActual(namespace string) (res map[string]*allocation.PodHeader) {
	s.mu.Lock()
	defer s.mu.Unlock()
	res = map[string]*allocation.PodHeader{}

	for _, what := range []map[string]*allocation.Pod{
		s.pending, s.active, s.ready,
	} {
		for k, v := range what {
			if _, ok := res[k]; !ok {
				if v == nil {
					res[k] = nil
					continue
				}
				res[k] = v.PodHeader
			}
		}
	}
	for k, v := range res {
		if v == nil || v.Namespace != namespace {
			delete(res, k)
		}
	}
	return
}

// returns latest (done/active/pending) pod
func (s *State) getLatest(name string) (res *allocation.Pod) {
	var ok bool
	if res, ok = s.pending[name]; ok {
		return
	}
	if res, ok = s.active[name]; ok {
		return
	}
	if res, ok = s.ready[name]; ok {
		return
	}
	return
}
