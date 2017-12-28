package provider

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"github.com/mitchellh/copystructure"
)

// Providers evaluator
type Evaluator struct {
	*supervisor.Control
	log       *logx.Log
	estimator Manager

	state map[string]allocation.ProviderSlice
	dirty map[string]struct{} // dirty

	allocateChan   chan *allocation.Pod
	deallocateChan chan string
}

func NewEvaluator(ctx context.Context, log *logx.Log, estimator Manager, state allocation.PodSlice) (e *Evaluator) {
	e = &Evaluator{
		Control:        supervisor.NewControl(ctx),
		log:            log.GetLog("provider", "evaluator"),
		estimator:      estimator,
		state:          map[string]allocation.ProviderSlice{},
		dirty:          map[string]struct{}{},
		allocateChan:   make(chan *allocation.Pod),
		deallocateChan: make(chan string),
	}
	for _, pod := range state {
		if pod.Providers != nil || len(pod.Providers) > 0 {
			v, _ := copystructure.Copy(pod.Providers)
			e.state[pod.Name] = v.(allocation.ProviderSlice)
			e.dirty[pod.Name] = struct{}{}
		}
	}
	return
}

func (e *Evaluator) Open() (err error) {
	go e.loop()
	err = e.Control.Open()
	return
}

// Returns base constraint from manifest. For pods without resources GetConstraint adds constraint "__provider.allocate = false".
func (e *Evaluator) GetConstraint(pod *manifest.Pod) manifest.Constraint {
	if pod.Providers == nil || len(pod.Providers) == 0 {
		return pod.Constraint.Merge(manifest.Constraint{
			"__provider.allocate": "= false",
		})
	}
	return pod.Constraint.Clone()
}

// Allocate providers in given pod
func (e *Evaluator) Allocate(pod *manifest.Pod, env map[string]string) {
	go func() {
		var alloc allocation.Pod
		if err := alloc.FromManifest(pod, env); err != nil {
			return
		}
		select {
		case <-e.Control.Ctx().Done():
			e.log.Errorf(`skip allocate %s: %v`, pod.Name, e.Control.Ctx().Err())
		case e.allocateChan <- &alloc:
			e.log.Tracef(`allocate: %s %d`, pod.Name, pod.Mark())
		}
	}()
}

// Deallocate all providers in given pod
func (e *Evaluator) Deallocate(name string) {
	go func() {
		select {
		case <-e.Control.Ctx().Done():
			e.log.Errorf(`skip deallocate %s: %v`, name, e.Control.Ctx().Err())
		case e.deallocateChan <- name:
			e.log.Tracef(`deallocate: %s`, name)
		}
	}()
}

func (e *Evaluator) loop() {
	log := e.log.WithTags("evaluator", "loop")
	log.Tracef("open")
LOOP:
	for {
		select {
		case <-e.Control.Ctx().Done():
			break LOOP
		case alloc := <-e.allocateChan:
			log.Tracef(`allocate: %s`, alloc.Name)
			e.evaluate(alloc.Name, alloc.Providers)
		case name := <-e.deallocateChan:
			log.Tracef(`deallocate: %s`, name)
			e.evaluate(name, nil)
		}
	}
	log.Tracef("close")
}

func (e *Evaluator) evaluate(name string, providers allocation.ProviderSlice) {
	var left allocation.ProviderSlice

	if _, isDirty := e.dirty[name]; !isDirty {
		left = e.state[name]
	}
	c, u, d := Plan(left, providers)
	e.updateState(name, providers)

	// send
	var id string
	for _, provider := range c {
		id = provider.GetID(name)
		e.estimator.CreateProvider(id, provider)
		e.log.Debugf(`create %s:%v sent to estimator`, id, provider)
	}
	for _, provider := range u {
		id = provider.GetID(name)
		e.estimator.UpdateProvider(id, provider)
		e.log.Debugf(`update %s:%v sent to estimator`, id, provider)
	}
	for _, prov := range d {
		id = name + "." + prov
		e.estimator.DestroyProvider(id)
		e.log.Debugf(`destroy %s sent to estimator`, id)
	}
}

func (e *Evaluator) updateState(name string, providers allocation.ProviderSlice) {
	if _, ok := e.dirty[name]; ok {
		delete(e.dirty, name)
		e.log.Debugf(`pod %s reached clean state`, name)
	}
	if providers == nil {
		delete(e.state, name)
		e.log.Tracef(`pod %s removed from state`, name)
		return
	}
	e.state[name] = providers
	e.log.Tracef(`pod %s updated in state`, name)
}
