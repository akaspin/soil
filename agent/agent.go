package agent

import (
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/manifest"
)

type Scheduler interface {
	// SyncNamespace internal state with given manifests
	Sync(namespace string, pods []*manifest.Pod) (err error)
}

type Source interface {
	Name() string

	Namespaces() []string

	Mark() bool

	Required() manifest.Constraint

	// Bind consumer. Source source will call callback on
	// change states.
	Register(callback func(active bool, env map[string]string))

	SubmitPod(name string, constraints manifest.Constraint)

	RemovePod(name string)
}

type Configurable interface {
	Set(v map[string]string, replace bool) (err error)
	Delete(keys ...string) (err error)
}

type AllocationReporter interface {

	// sync
	Sync(pods []*allocation.Pod)

	// report about allocation
	Report(name string, pod *allocation.Pod, failures []error)
}

