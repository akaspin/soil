package cluster

import "github.com/akaspin/soil/agent/bus"

const (
	backendNone   = "none"
	backendConsul = "consul"
)

// KV Worker
type Worker interface {
	Store(message bus.Message)
	StoreTTL(message bus.Message)
}
