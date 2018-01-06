package allocation

import (
	"fmt"
)

// Allocations state
type PodSlice []*Pod

func (s *PodSlice) FromFilesystem(systemPaths SystemPaths, discoveryFunc func() ([]string, error)) (err error) {
	paths, err := discoveryFunc()
	var failures []error
	for _, path := range paths {
		pod := NewPod(systemPaths)
		if parseErr := pod.FromFilesystem(path); parseErr != nil {
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
