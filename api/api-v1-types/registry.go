package api_v1_types

import "github.com/akaspin/soil/manifest"

type RegistrySubmitRequest manifest.Pods
type RegistrySubmitResponse map[string]uint64

type RegistryGetResponse manifest.Pods
