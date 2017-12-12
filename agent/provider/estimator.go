package provider

import "github.com/akaspin/soil/agent/allocation"

// Estimator can reconfigure resources from providers
type Estimator interface {

	// Create provider
	CreateProvider(id string, alloc *allocation.Provider)

	// Update provider
	UpdateProvider(id string, alloc *allocation.Provider)

	// Deallocate provider on estimator
	DestroyProvider(id string)
}
