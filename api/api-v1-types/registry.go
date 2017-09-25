package api_v1_types

import "github.com/akaspin/soil/manifest"

type RegistryGetResponse map[string]manifest.Registry

type RegistryPutRequest manifest.Registry
type RegistryPutResponse map[string]uint64

