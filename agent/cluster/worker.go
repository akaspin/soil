package cluster

import (
	"context"
	"github.com/akaspin/soil/agent/bus"
	"io"
	"time"
)

const (
	backendNone   = "none"
	backendLocal  = "local"
	backendConsul = "consul"
)

type WorkerConfig struct {
	Kind    string
	ID      string
	Address string
	TTL     time.Duration
}

// KV Worker
type Worker interface {
	io.Closer

	// Submit operations
	Submit(op []WorkerStoreOp)

	// Clean context closes then worker is ready to accept operations
	CleanCtx() context.Context

	// Failure context closes then worker is failed and needs to be recreated
	FailureCtx() context.Context

	CommitChan() chan []string
}

type WorkerStoreOp struct {
	Message bus.Message
	WithTTL bool
}
