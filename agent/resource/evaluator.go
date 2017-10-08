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
	log          *logx.Log
	workerConfig EvaluatorConfig

	consumer bus.MessageConsumer

	workers map[string]*Worker // worker

	configChan     chan []Config
	allocateChan   chan *manifest.Pod
	deallocateChan chan string
}

func NewEvaluator(ctx context.Context, log *logx.Log, workerConfig EvaluatorConfig, state allocation.Recovery, consumer bus.MessageConsumer) (e *Evaluator) {
	e = &Evaluator{
		Control:      supervisor.NewControl(ctx),
		log:          log.GetLog("resource", "evaluator"),
		workerConfig: workerConfig,

		consumer: consumer,
		workers:  map[string]*Worker{},

		configChan:     make(chan []Config),
		allocateChan:   make(chan *manifest.Pod),
		deallocateChan: make(chan string),
	}
	byType := map[string][]*Allocation{}
	for _, alloc := range state {
		for _, r := range alloc.Resources {
			rType := r.Request.Kind
			byType[rType] = append(byType[rType], &Allocation{
				PodName:  alloc.Name,
				Resource: r,
			})
		}
	}
	for name, a := range byType {
		e.workers[name] = NewWorker(e.Control.Ctx(), e.log, name, e.workerConfig, a)
		e.log.Debugf(`worker "%s" created: %v`, name, a)
	}
	return
}

func (e *Evaluator) Open() (err error) {
	go e.loop()
	err = e.Control.Open()
	return
}

func (e *Evaluator) Configure(configs ...Config) {
	select {
	case <-e.Control.Ctx().Done():
	case e.configChan <- configs:
	}
}

func (e *Evaluator) GetConstraint(pod *manifest.Pod) manifest.Constraint {
	return pod.GetResourceRequestConstraint()
}

func (e *Evaluator) Allocate(pod *manifest.Pod, env map[string]string) {
	select {
	case <-e.Control.Ctx().Done():
	case e.allocateChan <- pod:
	}
}

func (e *Evaluator) Deallocate(name string) {
	select {
	case <-e.Control.Ctx().Done():
	case e.deallocateChan <- name:
	}
}

func (e *Evaluator) loop() {
	for {
		select {
		case <-e.Control.Ctx().Done():
			return
		case configs := <-e.configChan:
			e.log.Tracef("received configure request: %v", configs)
			e.handleConfigs(configs)
		case ingest := <-e.allocateChan:
			e.log.Tracef("received allocate request: %v", ingest)
			e.handleAlloc(ingest.Name, ingest)
		case ingest := <-e.deallocateChan:
			e.log.Tracef("received deallocate request: %s", ingest)
			e.handleAlloc(ingest, nil)
		}
	}
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
			e.log.Infof(`removed worker: %s`, k)
		}
	}

	for name, config := range byName {
		if w, ok := e.workers[name]; ok {
			e.log.Debugf(`sending config to worker "%s": %v"`, name, config)
			w.Configure(config)
			continue
		}

		e.workers[name] = NewWorker(e.Control.Ctx(), e.log, name, EvaluatorConfig{}, nil)
		e.workers[name].Configure(config)
		e.log.Infof(`worker created "%s": %v"`, name, config)
	}
}

func (e *Evaluator) handleAlloc(name string, pod *manifest.Pod) {
	byType := map[string][]*manifest.Resource{}
	if pod != nil {
		for _, r := range pod.Resources {
			byType[r.Kind] = append(byType[r.Kind], r)
		}
	}
	for k, v := range e.workers {
		r := byType[k]
		e.log.Debugf(`sending requests to worker "%s": %s:%v`, k, name, r)
		v.Submit(name, r)
	}
}
