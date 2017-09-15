package agent

import (
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/manifest"
)

type Scheduler interface {
	// Sync specific namespace with given manifests
	Sync(namespace string, pods []*manifest.Pod) (err error)
}

type EvaluationReporter interface {
	Sync(pods []*allocation.Pod)
	Report(name string, pod *allocation.Pod, failures []error)
}
