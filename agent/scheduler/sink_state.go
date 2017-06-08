package scheduler

import (
	"github.com/akaspin/soil/manifest"
	"sync"
)

type SinkState struct {
	// dirty pods from executor
	dirty         map[string]string
	namespaces    []string
	registrations map[string]map[string]*manifest.Pod
	mu            *sync.Mutex
}

func NewSinkState(namespaces []string, dirty map[string]string) (s *SinkState) {
	s = &SinkState{
		dirty:         dirty,
		namespaces:    namespaces,
		registrations: map[string]map[string]*manifest.Pod{},
		mu:            &sync.Mutex{},
	}
	for _, n := range namespaces {
		s.registrations[n] = map[string]*manifest.Pod{}
	}
	return
}

// SyncNamespace syncs pods definitions in specific namespace.
// Returns actual changes
func (s *SinkState) SyncNamespace(namespace string, pods []*manifest.Pod) (changes map[string]*manifest.Pod) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var ok bool

	// Save current state
	current := map[string]*manifest.Pod{}
	for _, ns := range s.namespaces {
		for name, pod := range s.registrations[ns] {
			if _, ok = current[name]; !ok {
				current[name] = pod
			}
		}
	}

	changes = map[string]*manifest.Pod{}
	ingest := map[string]*manifest.Pod{}
	for _, pod := range pods {
		ingest[pod.Name] = pod
	}

	// evaluate deletions from given namespace
	for name := range s.registrations[namespace] {
		var in *manifest.Pod
		if in, ok = ingest[name]; !ok {
			delete(s.registrations[namespace], name)
		} else {
			s.registrations[in.Namespace][in.Name] = in
		}
		changes[name] = s.get(name)
		delete(s.dirty, name)
	}
	// evaluate deletions from dirty stale
	for name, ns := range s.dirty {
		if ns == namespace {
			if _, ok := ingest[name]; !ok {
				delete(s.dirty, name)
				changes[name] = nil
			}
		}
	}
	// add all other pods
	s.registrations[namespace] = ingest
	for _, ns := range s.namespaces {
		for name, pod := range s.registrations[ns] {
			if _, ok = changes[name]; !ok {
				changes[name] = pod
			}
			delete(s.dirty, name)
		}
	}
	// cleanup
	for name, pod := range changes {
		if pod != nil && manifest.IsEqual(pod, current[name]) {
			delete(changes, name)
		}
	}

	return
}

func (s *SinkState) get(name string) (res *manifest.Pod) {
	var ok bool
	for _, namespace := range s.namespaces {
		if res, ok = s.registrations[namespace][name]; ok {
			return
		}
	}
	return
}
