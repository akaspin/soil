package resource

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"io"
)

const (
	dummyExecutorNature = "dummy"
	rangeExecutorNature = "range"
)

// Executor
type Executor interface {
	io.Closer

	// allocate resource
	Allocate(request Alloc)
	Deallocate(id string)
}

// replaceable Executor instance
type ExecutorInstance struct {
	ctx      context.Context
	cancel   context.CancelFunc
	log      *logx.Log
	consumer bus.Consumer

	ExecutorConfig Config
	Executor       Executor
}

func NewExecutorInstance(ctx context.Context, log *logx.Log, evaluatorConfig EvaluatorConfig, executorConfig Config, consumer bus.Consumer) (i *ExecutorInstance, err error) {
	i = &ExecutorInstance{
		log:            log.GetLog("resource", "executor", "instance", executorConfig.Nature, executorConfig.Kind),
		ExecutorConfig: executorConfig,
		consumer:       consumer,
	}
	i.ctx, i.cancel = context.WithCancel(ctx)

	executorLog := log.GetLog("resource", "worker", executorConfig.Kind, executorConfig.Nature)
	switch executorConfig.Nature {
	case dummyExecutorNature:
		i.Executor = NewDummyExecutor(executorLog, executorConfig, i)
	case rangeExecutorNature:
		i.Executor = NewRangeExecutor(executorLog, executorConfig, i)
	default:
		err = fmt.Errorf("unknown Executor nature: %v", executorConfig)
	}
	return
}

func (i *ExecutorInstance) Close() (err error) {
	i.cancel()
	return
}

func (i *ExecutorInstance) ConsumerName() string {
	return i.ExecutorConfig.Nature + "." + i.ExecutorConfig.Kind
}

func (i *ExecutorInstance) ConsumeMessage(message bus.Message) {
	go func() {
		select {
		case <-i.ctx.Done():
			i.log.Tracef("ignoring %v: %v", message, i.ctx.Err())
		default:
			i.consumer.ConsumeMessage(message)
		}
	}()
}

func NewExecutorMessage(id string, err error, values map[string]string) (res bus.Message) {
	if err != nil {
		res = bus.NewMessage(id, map[string]string{
			"allocated": "false",
			"failure":   fmt.Sprint(err),
		})
		return
	}
	payload := map[string]string{
		"allocated": "true",
	}
	for k, v := range values {
		payload[k] = v
	}
	res = bus.NewMessage(id, payload)
	return
}
