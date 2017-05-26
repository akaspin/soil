package agent

import (
	"github.com/akaspin/soil/manifest"
)

type Scheduler interface {
	// SyncNamespace internal state with given manifests
	Sync(namespace string, pods []*manifest.Pod) (err error)
}

// Arbiter holds any state and returns internal values
type Arbiter interface {
	// Name returns arbiter name
	Name() string

	RegisterManager(callback func(env map[string]string)) (current map[string]string, marked bool)

	// RegisterPod returns values for given fields. Arbiter may
	// evaluate given fields. For example try to allocate counter.
	RegisterPod(name string, fields []string)

	// DeregisterPod pod
	DeregisterPod(name string)
}