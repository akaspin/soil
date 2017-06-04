package scheduler

import (
	"context"
	"github.com/akaspin/concurrency"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/supervisor"
)

func New(ctx context.Context, log *logx.Log, workers int, arbiter ...agent.Source) (sink *Sink, sv supervisor.Component) {
	pool := concurrency.NewWorkerPool(ctx, concurrency.Config{
		Capacity: workers,
	})
	executor := NewExecutor(ctx, log, pool)
	manager := NewArbiter(ctx, log, arbiter...)
	var arbiterComponents []supervisor.Component
	for _, a := range arbiter {
		arbiterComponents = append(arbiterComponents, a.(supervisor.Component))
	}
	sink = NewSink(ctx, log, executor, manager)
	sv = supervisor.NewChain(ctx,
		supervisor.NewGroup(ctx,
			supervisor.NewChain(ctx,
				pool,
				executor,
			),
			supervisor.NewChain(ctx,
				supervisor.NewGroup(ctx, arbiterComponents...),
				manager,
			)),
		sink,
	)
	return
}
