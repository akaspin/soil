package scheduler

import (
	"github.com/akaspin/soil/manifest"
)

type Evaluator interface {
	// Get constraint
	GetConstraint(pod *manifest.Pod) manifest.Constraint

	// Function to invoke then pod constraint is passed.
	// Allocate should be thread-safe and non-blocking
	Allocate(name string, pod *manifest.Pod, env map[string]string)

	// Function to invoke then pod constraint is failed.
	// Deallocate should be thread-safe and non-blocking
	Deallocate(name string)
}

type ManagedEvaluator struct {
	manager   *Manager
	evaluator Evaluator
}

func NewManagedEvaluator(manager *Manager, evaluator Evaluator) (m ManagedEvaluator) {
	m = ManagedEvaluator{
		manager:   manager,
		evaluator: evaluator,
	}
	return
}
