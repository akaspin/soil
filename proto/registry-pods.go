package proto

import "github.com/akaspin/soil/manifest"

type RegistryPodsContents map[string]manifest.Registry

type RegistryPodsPutRequest manifest.Registry
type RegistryPodsDeleteRequest []string
