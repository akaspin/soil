package cluster

import (
	"context"
	"github.com/akaspin/logx"
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

	commitsChan      chan []StoreCommit
	watchResultsChan chan WatchResult
}

func newBaseBackend(ctx context.Context, log *logx.Log, config BackendConfig) (b *baseBackend) {
	b = &baseBackend{
		log:              log,
		config:           config,
		commitsChan:      make(chan []StoreCommit, 1),
		watchResultsChan: make(chan WatchResult, 1),
	}
	b.failCtx, b.failCancel = context.WithCancel(context.Background())
	b.ctx, b.cancel = context.WithCancel(b.failCtx)
	b.readyCtx, b.readyCancel = context.WithCancel(context.Background())
	// also close local context on close parent
	go func() {
		<-ctx.Done()
		b.cancel()
	}()
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

func (b *baseBackend) CommitChan() chan []StoreCommit {
	return b.commitsChan
}

func (b *baseBackend) WatchResultsChan() chan WatchResult {
	return b.watchResultsChan
}

func (b *baseBackend) fail(err error) {
	b.log.Error(err)
	b.failCancel()
}
