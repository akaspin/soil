package resource

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/mitchellh/hashstructure"
	"io"
)

const (
	dummyExecutorNature = "dummy"
	rangeExecutorNature = "range"
)

// ExecutorConfig represents one resource in Agent configuration
type ExecutorConfig struct {
	Nature     string                 // Worker nature
	Kind       string                 // Declared type
	Properties map[string]interface{} // Properties
}

func (c *ExecutorConfig) IsEqual(config ExecutorConfig) (res bool) {
	leftHash, _ := hashstructure.Hash(*c, nil)
	rightHash, _ := hashstructure.Hash(config, nil)
	res = leftHash == rightHash
	return
}

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
	consumer bus.MessageConsumer

	ExecutorConfig ExecutorConfig
	Executor       Executor
}

func NewExecutorInstance(ctx context.Context, log *logx.Log, evaluatorConfig EvaluatorConfig, executorConfig ExecutorConfig, consumer bus.MessageConsumer) (i *ExecutorInstance, err error) {
	i = &ExecutorInstance{
		log:            log.GetLog("resource", "executor", "instance", executorConfig.Nature, executorConfig.Kind),
		ExecutorConfig: executorConfig,
		consumer:       consumer,
	}
	i.ctx, i.cancel = context.WithCancel(ctx)

	switch executorConfig.Nature {
	case dummyExecutorNature:
		i.Executor = NewDummyExecutor(log.GetLog("resource", "worker", executorConfig.Kind, executorConfig.Nature),
			executorConfig, i)
	default:
		err = fmt.Errorf("unknown Executor nature: %v", executorConfig)
	}
	return
}

func (i *ExecutorInstance) Close() (err error) {
	i.cancel()
	return
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
			"failure": fmt.Sprint(err),
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