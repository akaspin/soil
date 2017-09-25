package registry

import "github.com/akaspin/soil/manifest"

type Consumer interface {

	ConsumeRegistry(namespace string, payload manifest.Registry)
}