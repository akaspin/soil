package allocation

import "fmt"

type StateHolder interface {
	GetState() State
}

// Allocations state
type State []*Pod

// Recover state from files
func (s *State) FromFS(systemPaths SystemPaths, paths ...string) (err error) {
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