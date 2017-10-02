package registry

import (
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/soil/agent/bus"
)

type Evaluator interface {
	// Get constraint
	GetConstraint(pod *manifest.Pod) manifest.Constraint

	// Function to invoke then pod constraint is passed
	Allocate(name string, pod *manifest.Pod, env map[string]string)

	// Function to invoke then pod constraint is failed
	Deallocate(name string)
}

type ManagedEvaluator struct {
	Manager   *bus.Manager
	Evaluator Evaluator
}

func NewManagedEvaluator(manager *bus.Manager, evaluator Evaluator) (m ManagedEvaluator) {
	m = ManagedEvaluator{
		Manager:   manager,
		Evaluator: evaluator,
	}
	return
}
