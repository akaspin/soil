package allocation

import (
	"fmt"
)

// Allocations state
type State []*Pod

func (s *State) Discover(systemPaths SystemPaths, discoveryFunc func() ([]string, error)) (err error) {
	paths, err := discoveryFunc()
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
	return
}

func (s State) Find(name string) (res *Header) {
	for _, alloc := range s {
		if alloc.Name == name {
			res = alloc.Header
			break
		}
	}
	return
}

type SystemPaths struct {
	Local   string
	Runtime string
}

func DefaultSystemPaths() SystemPaths {
	return SystemPaths{
		Local:   dirSystemDLocal,
		Runtime: dirSystemDRuntime,
	}
}