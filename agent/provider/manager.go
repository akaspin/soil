package provider

import "github.com/akaspin/soil/agent/allocation"

// Manager can reconfigure resources from providers
type Manager interface {

	// Create provider
	CreateProvider(id string, alloc *allocation.Provider)

	// Update provider
	UpdateProvider(id string, alloc *allocation.Provider)

	// Deallocate provider on estimator
	DestroyProvider(id string)
}
