package resource

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/manifest"
)


type Worker struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	log        *logx.Log
	name       string
	evaluatorConfig EvaluatorConfig

	executor Executor
	config   *Config

	state map[string]*Allocation // key: pod.resource-kind
	dirty map[string]struct{}

	configChan chan Config
	valuesChan chan bus.Message
}

// Create new worker with recovered allocations
func NewWorker(ctx context.Context, log *logx.Log, name string, config EvaluatorConfig, recovered []*Allocation) (w *Worker) {
	w = &Worker{
		log:        log.GetLog("resource", "worker", name),
		name: name,
		evaluatorConfig: config,
		state:      map[string]*Allocation{},
		dirty:      map[string]struct{}{},
		configChan: make(chan Config, 1),
		valuesChan: make(chan bus.Message, 1),
	}
	w.ctx, w.cancelFunc = context.WithCancel(ctx)
	for _, alloc := range recovered {
		w.state[alloc.GetId()] = alloc.Clone()
		w.dirty[alloc.GetId()] = struct{}{}
		w.log.Debugf("recovered: %s", alloc.GetId())
	}
	go w.loop()

	return
}

// Configure worker
func (w *Worker) Configure(config Config) {
	w.log.Tracef("configure: %v", config)
	select {
	case <-w.ctx.Done():
	case w.configChan <- config:
	}
}

// Allocate resources by specific pod
func (w *Worker) Submit(podName string, requests []*manifest.Resource) {
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
	if w.executor != nil {
		w.executor.Close()
	}
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
			log.Tracef(`received config: %v`, config)
			w.handleConfig(config)
		case message := <-w.valuesChan:
			log.Tracef(`received message: %v`, message)
			w.handleMessage(message)
		}
	}
	log.Debugf("close")
}

func (w *Worker) handleConfig(config Config) {
	w.log.Tracef("config: %v", config)
	if w.config == nil || !w.config.IsEqual(&config) {
		w.log.Debugf("configure: %v->%v", w.config, &config)
		w.config = &config
		if w.executor != nil {
			w.log.Tracef("closing executor")
			w.executor.Close()
		}
		// mark all as dirty
		for k := range w.state {
			w.dirty[k] = struct{}{}
		}
		var err error
		if w.executor, err = NewExecutor(w.ctx, w.log, w.evaluatorConfig, config, w); err != nil {
			w.log.Error(err)
			return
		}
		for _, v := range w.state {
			go w.executor.Allocate(v.Clone())
		}
		w.log.Debugf("executor created")
	}
}

func (w *Worker) handleMessage(message bus.Message) {
	w.log.Tracef("message: %v", message)
	id := message.GetPrefix()
	delete(w.dirty, message.GetPrefix())
	if alloc, ok := w.state[id]; ok {
		payload := message.GetPayload()
		alloc.Values = payload
		w.log.Tracef("updated: %s:%v", id, payload)
	}
}
