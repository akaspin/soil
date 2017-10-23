package scheduler

import "github.com/akaspin/soil/manifest"

type RegistryConsumer interface {
	ConsumeRegistry(payload manifest.Registry)
}
