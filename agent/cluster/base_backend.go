package cluster

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
)

type baseBackend struct {
	log    *logx.Log
	config BackendConfig

	ctx         context.Context
	cancel      context.CancelFunc
	readyCtx    context.Context
	readyCancel context.CancelFunc
	failCtx     context.Context
	failCancel  context.CancelFunc

	commitsChan chan []BackendCommit
	watchChan   chan bus.Message
}

func newBaseBackend(ctx context.Context, log *logx.Log, config BackendConfig) (b *baseBackend) {
	b = &baseBackend{
		log:         log,
		config:      config,
		commitsChan: make(chan []BackendCommit, 1),
		watchChan:   make(chan bus.Message, 1),
	}
	b.failCtx, b.failCancel = context.WithCancel(ctx)
	b.ctx, b.cancel = context.WithCancel(b.failCtx)
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

func (b *baseBackend) FailCtx() context.Context {
	return b.failCtx
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

func (b *baseBackend) fail(err error) {
	b.log.Error(err)
	b.failCancel()
}
