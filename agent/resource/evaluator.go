package resource

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
)

// Resource evaluator
type Evaluator struct {
	*supervisor.Control
	log             *logx.Log
	evaluatorConfig EvaluatorConfig

	downstreamConsumer bus.Consumer // (resource)
	upstreamConsumer   bus.Consumer // (__resource.request.<kind>)

	workers      map[string]*Worker
	dirtyWorkers map[string]struct{}
	cache        map[string]bus.Message

	configChan     chan []Config
	allocateChan   chan *manifest.Pod
	deallocateChan chan string
	messageChan    chan bus.Message
}

func NewEvaluator(ctx context.Context, log *logx.Log, workerConfig EvaluatorConfig, state allocation.Recovery, downstream, upstream bus.Consumer) (e *Evaluator) {
	e = &Evaluator{
		Control:            supervisor.NewControl(ctx),
		log:                log.GetLog("resource", "evaluator"),
		evaluatorConfig:    workerConfig,
		downstreamConsumer: downstream,
		upstreamConsumer:   upstream,

		workers:      map[string]*Worker{},
		dirtyWorkers: map[string]struct{}{},
		cache:        map[string]bus.Message{},

		configChan:     make(chan []Config),
		allocateChan:   make(chan *manifest.Pod),
		deallocateChan: make(chan string),
		messageChan:    make(chan bus.Message),
	}
	byKind := map[string][]Alloc{}
	for _, alloc := range state {
		for _, r := range alloc.Resources {
			rKind := r.Request.Kind
			byKind[rKind] = append(byKind[rKind], Alloc{
				PodName: alloc.Name,
				Request: r.Request.Clone(),
				Values:  bus.NewMessage(r.Request.GetID(alloc.Name), r.Values),
			})
		}
	}
	for name, a := range byKind {
		e.workers[name] = NewWorker(e.Control.Ctx(), e.log, name, e, e.evaluatorConfig, a)
		e.dirtyWorkers[name] = struct{}{}
		e.log.Debugf(`worker "%s" created: %v`, name, a)
	}
	return
}

func (e *Evaluator) Open() (err error) {
	go e.loop()
	err = e.Control.Open()
	return
}

func (e *Evaluator) GetConstraint(pod *manifest.Pod) manifest.Constraint {
	return pod.GetResourceRequestConstraint()
}

func (e *Evaluator) Allocate(pod *manifest.Pod, env map[string]string) {
	go func() {
		select {
		case <-e.Control.Ctx().Done():
			e.log.Warningf(`ignore allocate "%s": %v`, pod.Name, e.Control.Ctx().Err())
		case e.allocateChan <- pod:
		}
	}()
}

func (e *Evaluator) Deallocate(name string) {
	go func() {
		select {
		case <-e.Control.Ctx().Done():
			e.log.Warningf(`ignore deallocate "%s": %v`, name, e.Control.Ctx().Err())
		case e.deallocateChan <- name:
		}
	}()
}

func (e *Evaluator) Configure(configs Configs) {
	select {
	case <-e.Control.Ctx().Done():
		e.log.Warningf(`ignore configs %v: %v`, configs, e.Control.Ctx().Err())
	case e.configChan <- configs:
	}
}

func (e *Evaluator) ConsumerName() string {
	return "resource-evaluator"
}

// Consume message from worker
func (e *Evaluator) ConsumeMessage(message bus.Message) {
	e.log.Tracef("message consumed: %v", message)
	go func() {
		select {
		case <-e.Control.Ctx().Done():
			e.log.Warningf(`ignore worker message %v: %v`, message, e.Control.Ctx().Err())
		case e.messageChan <- message:
		}
	}()
}

func (e *Evaluator) loop() {
	log := e.log.GetLog(e.log.Prefix(), append(e.log.Tags(), "loop")...)
	for {
		select {
		case <-e.Control.Ctx().Done():
			return
		case configs := <-e.configChan:
			log.Tracef("config: %v", configs)
			e.handleConfigs(configs)
		case req := <-e.allocateChan:
			log.Tracef("allocate: %v", req)
			e.handleAlloc(req.Name, req)
		case req := <-e.deallocateChan:
			log.Tracef("deallocate: %s", req)
			e.handleAlloc(req, nil)
		case message := <-e.messageChan:
			log.Tracef("message: %v", message)
			e.handleMessage(message)
		}
	}
}

func (e *Evaluator) handleMessage(message bus.Message) {
	kind := message.GetID()
	if _, ok := e.workers[kind]; !ok {
		e.log.Warningf(`ignore %v: worker "%s" not found`, message, kind)
		return
	}
	delete(e.dirtyWorkers, kind)
	e.cache[kind] = message
	e.notify()
}

func (e *Evaluator) notify() {
	if len(e.dirtyWorkers) > 0 {
		e.log.Tracef(`skip update: %d workers are dirty`, len(e.dirtyWorkers))
		return
	}
	upstream := map[string]string{}
	downstream := map[string]string{}
	for k, v := range e.cache {
		var chunk map[string]string
		if err := v.Payload().Unmarshal(&chunk); err != nil {
			e.log.Error(err)
			continue
		}
		upstream[`request.`+k+`.allow`] = "true"
		for item, value := range chunk {
			downstream[k+"."+item] = value
		}
	}
	e.upstreamConsumer.ConsumeMessage(bus.NewMessage("resource", upstream))
	e.downstreamConsumer.ConsumeMessage(bus.NewMessage("resource", downstream))
}

func (e *Evaluator) handleConfigs(configs []Config) {
	byName := map[string]Config{}
	for _, c := range configs {
		byName[c.Kind] = c
	}
	for k, w := range e.workers {
		if _, ok := byName[k]; !ok {
			w.Close()
			delete(e.workers, k)
			delete(e.dirtyWorkers, k)
			delete(e.cache, k)
			e.log.Infof(`removed worker: %s`, k)
		}
	}
	for name, config := range byName {
		if w, ok := e.workers[name]; ok {
			e.log.Debugf(`sending config to worker "%s": %v"`, name, config)
			w.Configure(config)
			continue
		}

		e.workers[name] = NewWorker(e.Control.Ctx(), e.log, name, e, EvaluatorConfig{}, nil)
		e.dirtyWorkers[name] = struct{}{}
		e.workers[name].Configure(config)
		e.log.Infof(`worker created "%s": %v"`, name, config)
	}
	e.notify()
}

func (e *Evaluator) handleAlloc(podName string, pod *manifest.Pod) {
	byKind := map[string][]manifest.Resource{}
	if pod != nil {
		for _, r := range pod.Resources {
			byKind[r.Kind] = append(byKind[r.Kind], r)
		}
	}
	for workerName, v := range e.workers {
		r := byKind[workerName]
		e.log.Debugf(`sending requests to worker "%s": %s:%v`, workerName, podName, r)
		v.Submit(podName, r)
	}
}
