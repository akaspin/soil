package agent

import (
	"github.com/akaspin/soil/manifest"
)

type Scheduler interface {
	// SyncNamespace internal state with given manifests
	Sync(namespace string, pods []*manifest.Pod) (err error)
}

type Source interface {
	// Name returns arbiter name
	Name() string

	// Bind consumer. Source source will call callback on
	// change states.
	Register(callback func(env map[string]string)) (env map[string]string, marked bool)


	SubmitPod(name string, constraints manifest.Constraint)

	RemovePod(name string)
}