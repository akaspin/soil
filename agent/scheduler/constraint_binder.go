package scheduler

import (
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/manifest"
)

// ConstraintBinder can bind and unbind specific function to specific callback
type ConstraintBinder interface {
	Bind(id string, constraint manifest.Constraint, callback func(reason error, message bus.Message))
	Unbind(id string, callback func())
}
