package allocation

import "fmt"

type StateHolder interface {
	GetState() State
}

// Allocations state
type State []*Pod

// Recover state from files
func (s *State) FromFS(systemPaths SystemDPaths, paths ...string) (err error) {
	var failures []error

	for _, path := range paths {
		pod := NewPod(systemPaths)
		if parseErr := pod.FromFS(path); parseErr != nil {
			failures = append(failures, parseErr)
			continue
		}
		*s = append(*s, pod)
	}
	if len(failures) > 0 {
		err = fmt.Errorf("%v", failures)
	}
	return
}