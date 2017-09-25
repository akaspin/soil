package agent

import (
	"github.com/akaspin/soil/agent/allocation"
)

type EvaluationReporter interface {
	Sync(pods []*allocation.Pod)
	Report(name string, pod *allocation.Pod, failures []error)
}
