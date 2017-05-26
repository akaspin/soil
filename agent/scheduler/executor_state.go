package scheduler

import (
	"github.com/pkg/errors"
	"sync"
)

var (
	AllocationNotFoundError = errors.New("allocation is not found")
	AllocationNotUnique     = errors.New("allocation is not unique")
)

type ExecutorState struct {
	ready   map[string]*Allocation
	active  map[string]*Allocation
	pending map[string]*Allocation

	mu      *sync.Mutex
}

func NewExecutorState(initial []*Allocation) (s *ExecutorState) {
	s = &ExecutorState{
		ready:   map[string]*Allocation{},
		active:  map[string]*Allocation{},
		pending: map[string]*Allocation{},
		mu:      &sync.Mutex{},
	}
	for _, a := range initial {
		s.ready[a.AllocationHeader.Name] = a
	}
	return
}

// Submit allocation to pending. Use <nil> for destroy.
// Submit returns ok if pods actually submitted.
func (s *ExecutorState) Submit(name string, pending *Allocation) (ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	latest := s.getLatest(name)
	if !s.comparePods(latest, pending) {
		s.pending[name] = pending
		ok = true
	}
	return
}

// Promote allocation from pending to active and return ready and active pair.
// or error if evaluation is not possible at this time.
func (s *ExecutorState) Promote(name string) (ready, active *Allocation, err error) {
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
func (s *ExecutorState) Commit(name string, failures []error) (destroyed bool, err error) {
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

// List
func (s *ExecutorState) ListActual() (res map[string]*AllocationHeader) {
	s.mu.Lock()
	defer s.mu.Unlock()
	res = map[string]*AllocationHeader{}

	for _, what := range []map[string]*Allocation{
		s.pending, s.active, s.ready,
	} {
		for k, v := range what {
			if _, ok := res[k]; !ok {
				if v == nil {
					res[k] = nil
					continue
				}
				res[k] = v.AllocationHeader
			}
		}
	}
	for k, v := range res {
		if v == nil {
			delete(res, k)
		}
	}
	return
}

// returns latest (done/active/pending) pod
func (s *ExecutorState) getLatest(name string) (res *Allocation) {
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

func (s *ExecutorState) comparePods(left, right *Allocation) (ok bool) {
	var leftMark, rightMark uint64
	if left != nil {
		leftMark = left.AllocationHeader.Mark()
	}
	if right != nil {
		rightMark = right.AllocationHeader.Mark()
	}
	ok = leftMark == rightMark
	return
}