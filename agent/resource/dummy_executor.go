package resource

import (
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
)

// Dummy Executor for testing purposes.
type DummyExecutor struct {
	log    *logx.Log
	config Config

	consumer bus.MessageConsumer
}

func NewDummyExecutor(log *logx.Log, config Config, consumer bus.MessageConsumer) (e *DummyExecutor) {
	e = &DummyExecutor{
		log:      log.GetLog("resource", "executor", config.Nature, config.Kind),
		config:   config,
		consumer: consumer,
	}
	return
}

func (e *DummyExecutor) Close() error {
	return nil
}

func (e *DummyExecutor) Allocate(request Alloc) {
	e.log.Debugf("allocate: %s %v %v", request.GetID(), request.Request.Config, request.Values)
	id := request.GetID()
	payload := map[string]string{
		"allocated": "true",
	}
	for k, v := range e.config.Properties {
		payload[k] = fmt.Sprint(v)
	}
	for k, v := range request.Request.Config {
		payload[k] = fmt.Sprint(v)
	}
	go e.consumer.ConsumeMessage(bus.NewMessage(id, payload))
}

func (e *DummyExecutor) Deallocate(id string) {
	e.log.Debugf("deallocate: %s", id)
	go e.consumer.ConsumeMessage(bus.NewMessage(id, nil))
}
