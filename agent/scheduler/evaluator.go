package scheduler

import (
	"github.com/akaspin/soil/manifest"
)

type Evaluator interface {
	// Get constraint
	GetConstraint(pod *manifest.Pod) manifest.Constraint

	// Function to invoke then pod constraint is passed.
	// Allocate should be thread-safe and non-blocking
	Allocate(pod *manifest.Pod, env map[string]string)

	// Function to invoke then pod constraint is failed.
	// Deallocate should be thread-safe and non-blocking
	Deallocate(name string)
}

type BoundedEvaluator struct {
	binder    ConstraintBinder
	evaluator Evaluator
}

func NewBoundedEvaluator(binder ConstraintBinder, evaluator Evaluator) (e BoundedEvaluator) {
	e = BoundedEvaluator{
		binder:    binder,
		evaluator: evaluator,
	}
	return
}
