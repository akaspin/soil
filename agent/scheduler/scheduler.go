package scheduler

import (
	"context"
	"github.com/akaspin/concurrency"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/supervisor"
)

// Returns new scheduler with supervisor chain
func New(ctx context.Context, log *logx.Log, workers int, sources []agent.Source, reporters []agent.AllocationReporter) (sink *Sink, sv supervisor.Component) {
	pool := concurrency.NewWorkerPool(ctx, concurrency.Config{
		Capacity: workers,
	})
	executor := NewEvaluator(ctx, log, pool, reporters...)
	arbiter := NewArbiter(ctx, log, sources...)
	sink = NewSink(ctx, log, executor, arbiter)

	sv = supervisor.NewChain(ctx,
		supervisor.NewGroup(ctx,
			supervisor.NewChain(ctx,
				pool,
				executor,
			),
			arbiter,
		),
		sink,
	)
	return
}
