package cluster

import (
	"context"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/logx"
)

type baseBackend struct {
	log    *logx.Log
	config BackendConfig
	ctx context.Context
	cancel context.CancelFunc
	readyCtx context.Context
	readyCancel context.CancelFunc
	commitsChan chan []BackendCommit
	watchChan chan bus.Message
}

func newBaseBackend(ctx context.Context, log *logx.Log, config BackendConfig) (b *baseBackend) {
	b = &baseBackend{
		log: log,
		config:config,
		commitsChan:       make(chan []BackendCommit, 1),
		watchChan:        make(chan bus.Message, 1),
	}
	b.ctx, b.cancel = context.WithCancel(ctx)
	b.readyCtx, b.readyCancel = context.WithCancel(context.Background())
	return
}

func (b *baseBackend) Close() error {
	b.cancel()
	return nil
}

func (b *baseBackend) Ctx() context.Context {
	return b.ctx
}

func (b *baseBackend) ReadyCtx() context.Context {
	return b.readyCtx
}

func (b *baseBackend) CommitChan() chan []BackendCommit {
	return b.commitsChan
}

func (b *baseBackend) WatchChan() chan bus.Message {
	return b.watchChan
}


