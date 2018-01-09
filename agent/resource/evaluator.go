package resource

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/bus/pipe"
	"github.com/akaspin/soil/agent/resource/estimator"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"regexp"
)

const (
	opProviderCreate = iota
	opProviderUpdate
	opProviderDestroy
)

type opProvider struct {
	op       int
	id       string
	provider *allocation.Provider
}

type Evaluator struct {
	*supervisor.Control
	log        *logx.Log
	upstream   bus.Consumer // upstream bus consumer
	downstream bus.Consumer // downstream consumer

	allocations map[string]allocation.ResourceSlice // allocations by pod
	sandboxes   map[string]*Sandbox

	providerOpChan chan opProvider

	allocateChan   chan *allocation.Pod
	deallocateChan chan string
}

func NewEvaluator(ctx context.Context, log *logx.Log, upstream, downstream bus.Consumer, dirty allocation.PodSlice) (e *Evaluator) {
	e = &Evaluator{
		Control:        supervisor.NewControl(ctx),
		log:            log.GetLog("resource", "evaluator"),
		upstream:       pipe.NewLift("provider", upstream),
		allocations:    map[string]allocation.ResourceSlice{},
		sandboxes:      map[string]*Sandbox{},
		providerOpChan: make(chan opProvider),
		allocateChan:   make(chan *allocation.Pod),
		deallocateChan: make(chan string),
	}
	e.downstream = pipe.NewFn(e.jsonPipeFn, pipe.NewLift("resource", downstream))
	for _, alloc := range dirty {
		e.allocations[alloc.Name] = alloc.Resources
		for _, res := range alloc.Resources {
			e.sandboxes[res.Request.Provider] = nil
		}
	}
	return
}

func (e *Evaluator) Open() (err error) {
	go e.loop()
	// reset upstream and downstream
	downstream := map[string]manifest.FlatMap{}
	for pod, resources := range e.allocations {
		for _, res := range resources {
			downstream[pod+"."+res.Request.Name] = res.Values
			if res.Values != nil {
				downstream[pod+"."+res.Request.Name] = res.Values.WithJSON("__values").Merge(manifest.FlatMap{
					"provider": res.Request.Provider,
				})
			}
		}
	}
	if err = e.downstream.(bus.Pipe).GetConsumer().ConsumeMessage(bus.NewMessage("", downstream)); err != nil {
		e.log.Error(err)
	}
	e.log.Debugf(`dirty state sent to downstream: %v`, downstream)
	upstream := map[string]manifest.FlatMap{}
	for k := range e.sandboxes {
		upstream[k] = manifest.FlatMap{
			"allocated": "true",
			"kind":      estimator.BlackholeEstimator,
		}
	}
	if err = e.upstream.ConsumeMessage(bus.NewMessage("", upstream)); err != nil {
		e.log.Error(err)
	}
	e.log.Debugf(`dirty state sent to upstream: %v`, upstream)

	// initialise dirty sandboxes
	for providerId := range e.sandboxes {
		e.sandboxes[providerId] = e.createSandbox(providerId, &allocation.Provider{
			Name: providerId,
			Kind: estimator.BlackholeEstimator,
		})
	}

	err = e.Control.Open()
	return
}

// If Pod has any resources returns "${provider.<resource-provider>.allocated}":"true"
// for each resource. Elsewhere returns "resource.allocate":"false".
func (e *Evaluator) GetConstraint(pod *manifest.Pod) (c manifest.Constraint) {
	if len(pod.Resources) == 0 {
		c = manifest.Constraint{
			"resource.allocate": "false",
		}
		return
	}
	c1 := manifest.Constraint{}
	for _, r := range pod.Resources {
		c1[fmt.Sprintf("${provider.%s.allocated}", r.Provider)] = "true"
	}
	c = pod.Constraint.Merge(c1)
	return
}

func (e *Evaluator) Allocate(pod *manifest.Pod, env map[string]string) {
	go func() {
		var alloc allocation.Pod
		if err := (&alloc).FromManifest(pod, env); err != nil {
			e.log.Error(err)
		}
		select {
		case <-e.Control.Ctx().Done():
			e.log.Warningf(`skip allocate "%s": %v`, pod, e.Control.Ctx().Err())
		case e.allocateChan <- &alloc:
			e.log.Tracef(`allocate sent: "%s"`, alloc)
		}
	}()
}

func (e *Evaluator) Deallocate(name string) {
	go func() {
		select {
		case <-e.Control.Ctx().Done():
			e.log.Warningf(`skip deallocate "%s": %v`, name, e.Control.Ctx().Err())
		case e.deallocateChan <- name:
			e.log.Tracef(`deallocate sent: "%s"`, name)
		}
	}()
}

// Create provider
func (e *Evaluator) CreateProvider(id string, alloc *allocation.Provider) {
	select {
	case <-e.Control.Ctx().Done():
		e.log.Warningf(`skip create "%s": %v`, id, e.Control.Ctx().Err())
	case e.providerOpChan <- opProvider{
		op:       opProviderCreate,
		id:       id,
		provider: alloc.Clone(),
	}:
		e.log.Tracef(`create sent: "%s":%v`, id, alloc)
	}
}

func (e *Evaluator) UpdateProvider(id string, alloc *allocation.Provider) {
	select {
	case <-e.Control.Ctx().Done():
		e.log.Warningf(`skip update "%s": %v`, id, e.Control.Ctx().Err())
	case e.providerOpChan <- opProvider{
		op:       opProviderUpdate,
		id:       id,
		provider: alloc.Clone(),
	}:
		e.log.Tracef(`update sent: "%s":%v`, id, alloc)
	}
}

func (e *Evaluator) DestroyProvider(id string) {
	select {
	case <-e.Control.Ctx().Done():
		e.log.Warningf(`skip destroy "%s": %v`, id, e.Control.Ctx().Err())
	case e.providerOpChan <- opProvider{
		op: opProviderDestroy,
		id: id,
	}:
		e.log.Tracef(`destroy sent: %s`, id)
	}
}

func (e *Evaluator) loop() {
LOOP:
	for {
		select {
		case <-e.Control.Ctx().Done():
			e.log.Trace(`closing`)
			break LOOP
		case op := <-e.providerOpChan:
			switch op.op {
			case opProviderCreate, opProviderUpdate:
				if sandbox, ok := e.sandboxes[op.id]; ok {
					if op.op == opProviderCreate {
						e.log.Warningf(`create provider "%s": already exists`, op.id)
					}
					sandbox.reconfigure(op.provider)
					e.log.Debugf(`configuration %v sent to provider "%s"`, op.provider, op.id)
					continue LOOP
				}
				if op.op == opProviderUpdate {
					e.log.Warningf(`update provider "%s": not found`, op.id)
				}
				e.sandboxes[op.id] = e.createSandbox(op.id, op.provider)
			case opProviderDestroy:
				if sandbox, ok := e.sandboxes[op.id]; ok {
					if err := sandbox.Shutdown(); err != nil {
						e.log.Error(err)
					}
					delete(e.sandboxes, op.id)
					e.log.Debugf(`destroyed sandbox %s`, op.id)
					continue LOOP
				}
				e.log.Warningf(`destroy provider "%s": not found`, op.id)
			}
		case alloc := <-e.allocateChan:
			pod := alloc.Name
			left := e.allocations[pod]
			c, u, d := Plan(left, alloc.Resources)
			e.log.Debugf(`pod "%s" planned: create:%v update:%v destroy:%v`, pod, c, u, d)
			e.allocations[pod] = alloc.Resources
			for _, req := range c {
				id := pod + "." + req.Request.Name
				if sandbox, ok := e.sandboxes[req.Request.Provider]; ok {
					sandbox.Create(id, req)
				}
			}
			for _, req := range u {
				id := pod + "." + req.Request.Name
				if sandbox, ok := e.sandboxes[req.Request.Provider]; ok {
					sandbox.Update(id, req)
				}
			}
			for _, req := range d {
				id := pod + "." + req.Request.Name
				if sandbox, ok := e.sandboxes[req.Request.Provider]; ok {
					sandbox.Destroy(id)
				}
			}
		case podName := <-e.deallocateChan:
			if resources, ok := e.allocations[podName]; ok {
				for _, resource := range resources {
					id := podName + "." + resource.Request.Name
					if sandbox, ok1 := e.sandboxes[resource.Request.Provider]; ok1 {
						sandbox.Destroy(id)
						continue
					}
					e.log.Warningf(`destroy "%s": provider "%s" not found`, id, resource.Request.Provider)
				}
				delete(e.allocations, podName)
				e.log.Infof(`deallocated "%s"`, podName)
				continue LOOP
			}
			e.log.Debugf(`deallocate pod "%s": not found`, podName)
		}
	}
}

func (e *Evaluator) createSandbox(id string, alloc *allocation.Provider) (s *Sandbox) {
	s = NewSandbox(
		SandboxConfig{
			GlobalConfig: estimator.GlobalConfig{},
			Ctx:          e.Control.Ctx(),
			Log:          e.log,
			Upstream:     e.upstream,
			Downstream:   e.downstream,
		},
		id,
		alloc)
	e.log.Tracef(`sandbox created: %s:%v`, id, alloc)

	// submit creation of all resources
	for pod, resources := range e.allocations {
		for _, res := range resources {
			if res.Request.Provider == id {
				resourceId := pod + "." + res.Request.Name
				s.Create(resourceId, res.Clone())
				e.log.Tracef(`initial resource %s:%v sent to provider "%s"`, resourceId, res, id)
			}
		}
	}
	return
}

func (e *Evaluator) jsonPipeFn(message bus.Message) (res bus.Message) {
	e.log.Tracef(`got message %s`, message)
	res = message
	if message.Payload().IsEmpty() {
		return
	}
	var payload manifest.FlatMap
	if err := message.Payload().Unmarshal(&payload); err != nil {
		e.log.Errorf(`failed to unmarshal %s: %v`, message, err)
		return
	}
	res = bus.NewMessage(message.Topic(), payload.Merge(payload.Filter(regexp.MustCompile(`provider`)).WithJSON(allocation.ResourceValuesPostfix)))
	return
}
