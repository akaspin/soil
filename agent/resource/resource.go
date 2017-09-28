package resource

import "github.com/akaspin/soil/agent/allocation"

type WorkerConfig struct {
	// From Agent configuration
	Properties map[string]interface{}

	// Initial state
	Allocated map[string]*allocation.Resource
}

type Worker interface {

	// Submit allocation request
	Submit(id string, params map[string]interface{})
}
