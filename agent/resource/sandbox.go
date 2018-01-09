package resource

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/resource/estimator"
	"github.com/akaspin/soil/manifest"
)

const (
	opResourceCreate = iota
	opResourceUpdate
	opResourceDestroy
)

type opResource struct {
	op       int
	id       string
	resource *allocation.Resource
}

type SandboxConfig struct {
	Ctx          context.Context
	Log          *logx.Log
	GlobalConfig estimator.GlobalConfig
	Upstream     bus.Consumer // "resource." upstream (lifted)
	Downstream   bus.Consumer // "provider." upstream (lifted)
}

// Sandbox manages underlying estimator and proxies notifications to downstream
type Sandbox struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	log        *logx.Log
	config     SandboxConfig
	id         string

	estimatorUuid string
	estimator     Estimator
	resources     map[string]*allocation.Resource //

	reconfigureChan chan *allocation.Provider
	shutdownChan    chan struct{}
	opChan          chan *opResource
	resultChan      chan *estimator.Result
}

func NewSandbox(config SandboxConfig, id string, provider *allocation.Provider) (s *Sandbox) {
	s = &Sandbox{
		log:             config.Log.GetLog("resource", "sandbox", id),
		config:          config,
		id:              id,
		resources:       map[string]*allocation.Resource{},
		reconfigureChan: make(chan *allocation.Provider),
		shutdownChan:    make(chan struct{}),
		opChan:          make(chan *opResource),
		resultChan:      make(chan *estimator.Result),
	}
	s.ctx, s.cancelFunc = context.WithCancel(config.Ctx)
	s.reconfigure(provider)
	go s.loop()
	return
}

// Update provider
func (s *Sandbox) Configure(p *allocation.Provider) (err error) {
	go func() {
		select {
		case <-s.ctx.Done():
			err = s.ctx.Err()
			s.log.Warningf(`skip reconfigure %v: %v`, p, err)
		case s.reconfigureChan <- p:
			s.log.Tracef(`provider configuration sent: %v`, p)
		}
	}()
	return
}

// Create resource with id
func (s *Sandbox) Create(id string, req *allocation.Resource) {
	go func() {
		select {
		case <-s.ctx.Done():
			err := s.ctx.Err()
			s.log.Warningf(`skip create %s:%v: %v`, id, req, err)
		case s.opChan <- &opResource{
			op:       opResourceCreate,
			id:       id,
			resource: req.Clone(),
		}:
			s.log.Tracef(`create sent: %s:%v`, id, req)
		}
	}()
	return
}

func (s *Sandbox) Update(id string, req *allocation.Resource) {
	go func() {
		select {
		case <-s.ctx.Done():
			err := s.ctx.Err()
			s.log.Warningf(`skip update %s:%v: %v`, id, req, err)
		case s.opChan <- &opResource{
			op:       opResourceUpdate,
			id:       id,
			resource: req.Clone(),
		}:
			s.log.Tracef(`update sent: %s:%v`, id, req)
		}
	}()
	return
}

func (s *Sandbox) Destroy(id string) {
	go func() {
		select {
		case <-s.ctx.Done():
			err := s.ctx.Err()
			s.log.Warningf(`skip destroy %s: %v`, id, err)
		case s.opChan <- &opResource{
			op: opResourceDestroy,
			id: id,
		}:
			s.log.Tracef(`destroy sent: %s`, id)
		}
	}()
	return
}

// Destroy all resources in sandbox and notify upstream and downstream
func (s *Sandbox) Shutdown() (err error) {
	close(s.shutdownChan)
	return
}

// close sandbox without deallocate any resources
func (s *Sandbox) Close() (err error) {
	s.cancelFunc()
	return
}

func (s *Sandbox) loop() {
	var err error
LOOP:
	for {
		select {
		case <-s.shutdownChan:
			s.estimator.Shutdown()
			s.cancelFunc()
			s.config.Upstream.ConsumeMessage(bus.NewMessage(s.id, nil))
			for id := range s.resources {
				s.config.Downstream.ConsumeMessage(bus.NewMessage(id, nil))
			}
			s.log.Info("shutdown complete")
			break LOOP
		case <-s.ctx.Done():
			s.log.Trace("closing")
			break LOOP
		case res := <-s.resultChan:
			if res.Uuid != s.estimatorUuid {
				s.log.Warningf(`ignore %s: outdated estimator %s(current) != %s(received)`, res.Message, s.estimatorUuid, res.Uuid)
				continue LOOP
			}
			s.log.Tracef(`received result %s`, res.Message)
			if res.Message.Payload().IsEmpty() {
				// empty: delete internal
				delete(s.resources, res.Message.Topic())
				s.config.Downstream.ConsumeMessage(res.Message)
				s.log.Infof(`destroyed: %s`, res.Message.Topic())
				continue LOOP
			}
			if state, ok := s.resources[res.Message.Topic()]; ok {
				var payload manifest.FlatMap
				if err = res.Message.Payload().Unmarshal(&payload); err != nil {
					s.log.Warning(err)
					continue LOOP
				}
				state.Values = payload
				s.config.Downstream.ConsumeMessage(bus.NewMessage(res.Message.Topic(), payload.Merge(manifest.FlatMap{
					"provider": s.id,
				})))
				continue LOOP
			}
			s.log.Errorf(`resource not found %s`, res.Message)
		case prov := <-s.reconfigureChan:
			s.log.Debugf(`configuration received: %v`, prov)
			if err = s.estimator.Close(); err != nil {
				s.log.Error(err)
			}
			s.reconfigure(prov)
		case op := <-s.opChan:
			switch op.op {
			case opResourceCreate:
				if _, ok := s.resources[op.id]; ok {
					s.log.Debugf(`create: resource already exists: %v`, op)
					continue LOOP
				}
				s.resources[op.id] = op.resource
				if err = s.estimator.Create(op.id, op.resource); err != nil {
					s.log.Error(err)
				}
			case opResourceUpdate:
				if _, ok := s.resources[op.id]; !ok {
					s.log.Debugf(`update: resource not found: %v`, op)
					continue LOOP
				}
				s.resources[op.id] = op.resource
				if err = s.estimator.Update(op.id, op.resource); err != nil {
					s.log.Error(err)
				}
			case opResourceDestroy:
				if _, ok := s.resources[op.id]; !ok {
					s.log.Debugf(`destroy: resource not found: %s`, op.id)
					continue LOOP
				}
				delete(s.resources, op.id)
				if s.estimator.Destroy(op.id); err != nil {
					s.log.Error(err)
				}
			}
		}
	}
	s.log.Info(`closed`)
}

func (s *Sandbox) reconfigure(p *allocation.Provider) {
	var err error
	s.estimator, err = GetEstimator(s.config.GlobalConfig, estimator.Config{
		Ctx:      s.ctx,
		Log:      s.log.GetLog("resource", "estimator", p.Kind, s.id),
		Provider: p,
		Id:       s.id,
	})
	if err != nil {
		s.log.Error(err)
	}
	var ch chan *estimator.Result
	var ctx context.Context
	s.estimatorUuid, ctx, ch = s.estimator.Results()
	go s.estimatorWatchDog(s.estimatorUuid, ctx, ch)

	for id, r := range s.resources {
		if err = s.estimator.Create(id, r.Clone()); err != nil {
			s.log.Error(err)
		}
		s.log.Tracef(`recovered %v sent to estimator %s`, r, s.estimatorUuid)
	}

	msg := bus.NewMessage(s.id, map[string]string{
		"allocated": "true",
		"kind":      p.Kind,
	})
	s.config.Upstream.ConsumeMessage(msg)
	s.log.Debugf(`upstream notified: %s`, msg)
}

//
func (s *Sandbox) estimatorWatchDog(uuid string, ctx context.Context, ch chan *estimator.Result) {
	log := s.log.GetLog(s.log.Prefix(), append(s.log.Tags(), "watchdog", uuid)...)
	log.Debug("open")
	for res := range ch {
		select {
		case <-s.ctx.Done():
			log.Tracef(`close: parent: %v`, uuid, s.ctx.Err())
			return
		case <-ctx.Done():
			log.Tracef(`close: %v`, uuid, ctx.Err())
			return
		case s.resultChan <- res:
			log.Debugf(`piped %s`, res.Message)
		}
	}
}
