package resource

import (
	"context"
	"encoding/json"
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
	consumer        bus.Consumer

	executorInstance *ExecutorInstance
	state            map[string]*Alloc // key: pod.resource-kind
	dirty            map[string]struct{}

	configChan  chan Config
	requestChan chan workerRequest
	valuesChan  chan bus.Message
}

// Create new worker with recovered allocations
func NewWorker(ctx context.Context, log *logx.Log, name string, consumer bus.Consumer, config EvaluatorConfig, recovered []Alloc) (w *Worker) {
	w = &Worker{
		log:             log.GetLog("resource", "worker", name),
		name:            name,
		evaluatorConfig: config,
		consumer:        consumer,

		// current allocations
		state: map[string]*Alloc{},
		// dirtyWorkers state
		dirty:       map[string]struct{}{},
		configChan:  make(chan Config, 1),
		requestChan: make(chan workerRequest, 1),
		valuesChan:  make(chan bus.Message, 1),
	}
	w.ctx, w.cancelFunc = context.WithCancel(ctx)
	for _, alloc := range recovered {
		cloned := alloc.Clone()
		w.state[cloned.GetID()] = &cloned
		w.dirty[cloned.GetID()] = struct{}{}
		w.log.Debugf("recovered: %v", cloned)
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

// Allocate resources for specific pod
func (w *Worker) Submit(podName string, requests []manifest.Resource) {
	w.log.Tracef("submit: %s:%v", podName, requests)
	select {
	case <-w.ctx.Done():
	case w.requestChan <- workerRequest{
		podName:  podName,
		requests: requests,
	}:
	}
}

// Consume message with values from worker. Message prefix should be resource id.
func (w *Worker) ConsumeMessage(message bus.Message) (err error) {
	w.log.Tracef(`message consumed: %v`, message)
	select {
	case <-w.ctx.Done():
	case w.valuesChan <- message:
	}
	return
}

func (w *Worker) Close() (err error) {
	w.cancelFunc()
	return
}

func (w *Worker) loop() {
	log := w.log.GetLog(w.log.Prefix(), append(w.log.Tags(), "loop")...)
	log.Trace("open")
LOOP:
	for {
		select {
		case <-w.ctx.Done():
			break LOOP
		case config := <-w.configChan:
			log.Tracef(`config: %v`, config)
			w.handleConfig(config)
		case req := <-w.requestChan:
			log.Tracef(`request %v`, req)
			w.handleRequest(req.podName, req.requests)
		case message := <-w.valuesChan:
			log.Tracef(`message: %v`, message)
			w.handleMessage(message)
		}
	}
	log.Trace("close")
}

func (w *Worker) handleConfig(config Config) {
	w.log.Tracef("config: %v", config)
	if w.executorInstance == nil || !w.executorInstance.ExecutorConfig.IsEqual(config) {
		w.log.Tracef("creating executor: %v", config)
		if w.executorInstance != nil {
			w.log.Tracef("closing executor instance")
			w.executorInstance.Close()
		}
		// mark all as dirtyWorkers
		for k := range w.state {
			w.dirty[k] = struct{}{}
		}
		var err error
		if w.executorInstance, err = NewExecutorInstance(w.ctx, w.log, w.evaluatorConfig, config, w); err != nil {
			w.log.Error(err)
			return
		}
		w.log.Debugf(`executor created: %v`, config)
		if len(w.state) > 0 {
			for _, v := range w.state {
				w.log.Tracef(`sending allocate request: %v`, v)
				w.executorInstance.Executor.Allocate((*v).Clone())
			}
		} else {
			w.notify()
		}
		return
	}
	w.log.Debugf("skip reconfigure: config is equal: %v", config)
}

func (w *Worker) handleRequest(podName string, requests []manifest.Resource) {
	names := map[string]struct{}{}
	for _, req := range requests {
		id := req.GetID(podName)
		names[id] = struct{}{}
		allocated, ok := w.state[id]
		if !ok || allocated.NeedChange(req) {
			// change
			w.dirty[id] = struct{}{}
			w.state[id] = &Alloc{
				PodName: podName,
				Request: req,
				Values:  bus.NewMessage(id, nil),
			}
			w.log.Debugf(`submitting allocate: %v`, w.state[id])
			w.executorInstance.Executor.Allocate((*w.state[id]).Clone())
		} else {
			w.log.Tracef(`skip submit %v: config is equal`, allocated)
		}
	}
	for id, allocated := range w.state {
		_, ok := names[id]
		if allocated.PodName == podName && !ok {
			w.log.Debugf(`submitting deallocate: %v`, allocated)
			w.executorInstance.Executor.Deallocate(id)
		}
	}
}

func (w *Worker) handleMessage(message bus.Message) {
	w.log.Tracef("message: %v", message)
	prefix := message.GetID()
	delete(w.dirty, prefix)
	if !message.Payload().IsEmpty() {
		var allocated *Alloc
		var ok bool
		if allocated, ok = w.state[message.GetID()]; !ok {
			w.log.Warningf("not found: %s", prefix)
			return
		}
		allocated.Values = message
		w.log.Debugf("updated: %v", message)
	} else {
		delete(w.state, prefix)
		w.log.Debugf("removed: %v", message)
	}
	w.notify()
}

func (w *Worker) notify() {
	if len(w.dirty) > 0 {
		w.log.Debugf("skipping update: state is dirty %v", w.dirty)
		return
	}
	var err error
	data := map[string]string{}
	for id, all := range w.state {
		var value interface{}
		if err = all.Values.Payload().Unmarshal(&value); err != nil {
			w.log.Error(err)
			continue
		}
		var dataJson []byte
		if dataJson, err = json.Marshal(value); err != nil {
			w.log.Error(err)
			continue
		}

		data[id+".__values"] = string(dataJson)

		var chunk map[string]string
		if err = all.Values.Payload().Unmarshal(&chunk); err != nil {
			w.log.Error(err)
			continue
		}
		for k, v := range chunk {
			data[id+"."+k] = v
		}
	}
	w.log.Debugf("notify: %v", data)
	w.consumer.ConsumeMessage(bus.NewMessage(w.name, data))
}

type workerRequest struct {
	podName  string
	requests []manifest.Resource
}
