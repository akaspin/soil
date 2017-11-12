package cluster

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"io"
	"net/url"
	"time"
)

const (
	backendLocal  = "local"
	backendConsul = "consul"
)

type BackendConfig struct {
	Kind    string
	ID      string
	Address string
	Chroot  string
	TTL     time.Duration
}

type Backend interface {
	io.Closer

	Ctx() context.Context        // Backend context closes on backend is not available to accept operations
	FailCtx() context.Context    // Fail context closes then backend is failed
	ReadyCtx() context.Context   // Ready context closes then backend is ready to accept operations
	Submit(ops []BackendStoreOp) // Submit ops to backend
	Subscribe(req []BackendWatchRequest)
	CommitChan() chan []BackendCommit
	WatchChan() chan bus.Message
}

type BackendFactory func(ctx context.Context, log *logx.Log, config Config) (c Backend, err error)

func DefaultBackendFactory(ctx context.Context, log *logx.Log, config Config) (c Backend, err error) {
	kvConfig := BackendConfig{
		Kind:    "local",
		Chroot:  "soil",
		ID:      config.ID,
		Address: "localhost",
		TTL:     config.TTL,
	}
	u, err := url.Parse(config.URL)
	if err != nil {
		log.Error(err)
	}
	if err == nil {
		kvConfig.Kind = u.Scheme
		kvConfig.Address = u.Host
		kvConfig.Chroot = u.Path
	}
	kvLog := log.GetLog("cluster", "backend", kvConfig.Kind)
	switch kvConfig.Kind {
	//case backendConsul:
	//	c = NewConsulBackend(ctx, kvLog, kvConfig)
	default:
		c = NewZeroBackend(ctx, kvLog)
	}
	return
}

type BackendStoreOp struct {
	Message bus.Message
	WithTTL bool
}

type BackendWatchRequest struct {
	Key string
	Ctx context.Context
}

type BackendCommit struct {
	ID      string
	Hash    uint64
	WithTTL bool
}
