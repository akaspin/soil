package scheduler

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/supervisor"
)

// Returns new scheduler with supervisor chain
func New(ctx context.Context, log *logx.Log, reporters []agent.EvaluationReporter) (sink *Sink, arbiter *Manager, sv supervisor.Component) {
	executor := NewEvaluator(ctx, log, reporters...)
	arbiter = NewManager(ctx, log)
	sink = NewSink(ctx, log, executor, arbiter)

	sv = supervisor.NewChain(ctx,
		supervisor.NewGroup(ctx,
			executor,
			arbiter,
		),
		sink,
	)
	return
}
