package resource

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/manifest"
)

type Worker struct {
	ctx             context.Context
	cancelFunc      context.CancelFunc
	log             *logx.Log
	name            string
	evaluatorConfig EvaluatorConfig
	consumer        bus.MessageConsumer

	executorInstance *ExecutorInstance
	state            map[string]*Alloc // key: pod.resource-kind
	dirty            map[string]struct{}

	configChan chan ExecutorConfig
	valuesChan chan bus.Message
}

// Create new worker with recovered allocations
func NewWorker(ctx context.Context, log *logx.Log, name string, config EvaluatorConfig, recovered []Alloc) (w *Worker) {
	w = &Worker{
		log:             log.GetLog("resource", "worker", name),
		name:            name,
		evaluatorConfig: config,

		// current allocations
		state: map[string]*Alloc{},
		// dirty state
		dirty:      map[string]struct{}{},
		configChan: make(chan ExecutorConfig, 1),
		valuesChan: make(chan bus.Message, 1),
	}
	w.ctx, w.cancelFunc = context.WithCancel(ctx)
	for _, alloc := range recovered {
		w.state[alloc.GetID()] = &alloc
		w.dirty[alloc.GetID()] = struct{}{}
		w.log.Debugf("recovered: %s", alloc.GetID())
	}
	go w.loop()

	return
}

// Configure worker
func (w *Worker) Configure(config ExecutorConfig) {
	w.log.Tracef("configure: %v", config)
	select {
	case <-w.ctx.Done():
	case w.configChan <- config:
	}
}

// Allocate resources for specific pod
func (w *Worker) Submit(podName string, requests []manifest.Resource) {
	w.log.Tracef("submit: %s:%v", podName, requests)

}

// Consume message with values from worker. Message prefix should be resource id.
// Empty message means that resource is deallocated.
func (w *Worker) ConsumeMessage(message bus.Message) {
	w.log.Tracef(`message consumed: %v`, message)
	select {
	case <-w.ctx.Done():
	case w.valuesChan <- message:
	}
}

func (w *Worker) Close() (err error) {
	w.cancelFunc()
	return
}

func (w *Worker) loop() {
	log := w.log.GetLog(w.log.Prefix(), append(w.log.Tags(), "loop")...)
	log.Debugf("open")
LOOP:
	for {
		select {
		case <-w.ctx.Done():
			break LOOP
		case config := <-w.configChan:
			log.Tracef(`received ExecutorConfig: %v`, config)
			w.handleConfig(config)
		case message := <-w.valuesChan:
			log.Tracef(`received message: %v`, message)
			w.handleMessage(message)
		}
	}
	log.Debugf("close")
}

func (w *Worker) handleConfig(config ExecutorConfig) {
	w.log.Tracef("config: %v", config)
	if w.executorInstance == nil || !w.executorInstance.ExecutorConfig.IsEqual(config) {
		w.log.Debugf("configure: %v", config)
		if w.executorInstance != nil {
			w.log.Tracef("closing executor instance")
			w.executorInstance.Close()
		}
		// mark all as dirty
		for k := range w.state {
			w.dirty[k] = struct{}{}
		}
		var err error
		if w.executorInstance, err = NewExecutorInstance(w.ctx, w.log, w.evaluatorConfig, config, w); err != nil {
			w.log.Error(err)
			return
		}
		for _, v := range w.state {
			w.executorInstance.Executor.Allocate((*v).Clone())
		}
		w.log.Debugf("Executor created")
	}
}

func (w *Worker) handleMessage(message bus.Message) {
	w.log.Tracef("message: %v", message)
	delete(w.dirty, message.GetPrefix())
	if alloc, ok := w.state[message.GetPrefix()]; ok {
		alloc.Values = message
		w.log.Tracef("updated: %v", message)
	}
}
