package resource

import (
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
)

// Dummy Executor for testing purposes.
type DummyExecutor struct {
	log      *logx.Log
	kind     string
	consumer bus.MessageConsumer
}

func NewDummyExecutor(log *logx.Log, kind string, consumer bus.MessageConsumer) (e *DummyExecutor) {
	e = &DummyExecutor{
		log:      log.GetLog("resource", "Executor", kind),
		kind:     kind,
		consumer: consumer,
	}
	return
}

func (e *DummyExecutor) Close() error {
	return nil
}

func (e *DummyExecutor) Allocate(request Alloc) {
	e.log.Tracef("allocate: %s %v %v", request.GetID(), request.Request.Config, request.Values)
	id := request.GetID()
	payload := map[string]string{
		"allocated": "true",
	}
	for k, v := range request.Request.Config {
		payload[k] = fmt.Sprint(v)
	}
	e.consumer.ConsumeMessage(bus.NewMessage(id, payload))
}

func (e *DummyExecutor) Deallocate(id string) {
	e.log.Tracef("deallocate: %s", id)
	e.consumer.ConsumeMessage(bus.NewMessage(id, nil))
}
