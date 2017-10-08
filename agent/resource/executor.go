package resource

import (
	"io"
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"fmt"
)

const (
	dummyExecutorType = "dummy"
)

// Executor
type Executor interface {
	io.Closer

	// allocate resource
	Allocate(request *Allocation)
	Deallocate(id string)
}

func NewExecutor(ctx context.Context, log *logx.Log, evaluatorConfig EvaluatorConfig, config Config, consumer bus.MessageConsumer) (e Executor, err error) {
	switch config.Nature {
	case dummyExecutorType:
		e = NewDummyExecutor(log, config.Kind, consumer)
	default:
		err = fmt.Errorf("unknown executor nature: %v", config)
	}
	return
}