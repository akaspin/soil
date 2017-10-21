package scheduler

import "github.com/akaspin/soil/manifest"

type RegistryConsumer interface {
	ConsumeRegistry(namespace string, payload manifest.Registry)
}
