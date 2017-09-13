package agent

import (
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/manifest"
)

type Scheduler interface {
	// Sync specific namespace with given manifests
	Sync(namespace string, pods []*manifest.Pod) (err error)
}

type Source interface {
	SourceProducer

	// pod namespaces managed by source
	Namespaces() []string

	// Is data used only in constraint or
	// available for interpolation
	Mark() bool
}

type Configurable interface {
	Set(v map[string]string, replace bool) (err error)
	Delete(keys ...string) (err error)
}

type EvaluationReporter interface {
	Sync(pods []*allocation.Pod)
	Report(name string, pod *allocation.Pod, failures []error)
}
