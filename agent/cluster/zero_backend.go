package cluster

import (
	"context"
	"github.com/akaspin/logx"
)

// Zero backend used for local purposes. Zero backend is never ready.
type ZeroBackend struct {
	*baseBackend
}

func NewZeroBackend(ctx context.Context, log *logx.Log) (w *ZeroBackend) {
	w = &ZeroBackend{
		baseBackend: newBaseBackend(ctx, log, BackendConfig{}),
	}
	return
}

func (w *ZeroBackend) Submit(ops []StoreOp) {
}

func (w *ZeroBackend) Subscribe(req []WatchRequest) {
}
