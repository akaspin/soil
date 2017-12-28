package estimator

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/manifest"
	"github.com/nu7hatch/gouuid"
)

const (
	opResourceCreate = iota
	opResourceUpdate
	opResourceDestroy
)

type resourceOp struct {
	op     int
	id     string
	config map[string]interface{}
	values map[string]string
}

type baseEngine interface {
	createFn(id string, config map[string]interface{}, values map[string]string) (res interface{}, err error)
	updateFn(id string, config map[string]interface{}) (res interface{}, err error)
	destroyFn(id string) (err error)
	shutdownFn() (err error)
}

// basic estimator
type base struct {
	ctx    context.Context
	cancel context.CancelFunc
	log    *logx.Log

	uuid         string
	globalConfig GlobalConfig
	config       Config
	engine       baseEngine

	resultChan   chan *Result
	opChan       chan *resourceOp
	shutdownChan chan struct{}
}

func newBase(globalConfig GlobalConfig, config Config, engine baseEngine) (b *base) {
	id1, _ := uuid.NewV4()

	b = &base{
		uuid:         id1.String(),
		log:          config.Log.GetLog(config.Log.Prefix(), append(config.Log.Tags(), id1.String())...),
		globalConfig: globalConfig,
		config:       config,
		engine:       engine,

		resultChan:   make(chan *Result, 1),
		opChan:       make(chan *resourceOp, 1),
		shutdownChan: make(chan struct{}),
	}
	b.ctx, b.cancel = context.WithCancel(config.Ctx)
	go b.loop()
	return
}

func (b *base) Close() error {
	b.cancel()
	return nil
}

func (b *base) Results() (id string, ctx context.Context, ch chan *Result) {
	id = b.uuid
	ctx = b.ctx
	ch = b.resultChan
	return
}

func (b *base) Create(id string, resource *allocation.Resource) (err error) {
	stub := resource.Clone()
	select {
	case <-b.ctx.Done():
		b.log.Warningf(`ignore create %s:%v: %v`, id, resource, b.ctx.Err())
		err = b.ctx.Err()
	case b.opChan <- &resourceOp{
		op:     opResourceCreate,
		id:     id,
		config: stub.Request.Config,
		values: stub.Values,
	}:
		b.log.Debugf(`accepted create: %s:%v`, id, resource)
	}
	return
}

func (b *base) Update(id string, resource *allocation.Resource) (err error) {
	stub := resource.Clone()
	select {
	case <-b.ctx.Done():
		b.log.Warningf(`ignore update %s:%v: %v`, id, resource, b.ctx.Err())
		err = b.ctx.Err()
	case b.opChan <- &resourceOp{
		op:     opResourceUpdate,
		id:     id,
		config: stub.Request.Config,
	}:
		b.log.Debugf(`accepted update: %s:%v`, id, resource)
	}
	return
}

func (b *base) Destroy(id string) (err error) {
	select {
	case <-b.ctx.Done():
		b.log.Warningf(`ignore destroy %s: %v`, id, b.ctx.Err())
		err = b.ctx.Err()
	case b.opChan <- &resourceOp{
		op: opResourceDestroy,
		id: id,
	}:
		b.log.Debugf(`accepted destroy: %s`, id)
	}
	return
}

func (b *base) Shutdown() {
	close(b.shutdownChan)
	return
}

func (b *base) send(id string, failure error, values manifest.FlatMap) {
	res := &Result{
		Uuid: b.uuid,
	}
	if failure != nil {
		res.Message = bus.NewMessage(id, map[string]string{
			"allocated": "false",
			"failure":   failure.Error(),
		})
	} else if values != nil {
		res.Message = bus.NewMessage(id, values.Merge(manifest.FlatMap{
			"allocated": "true",
		}))
	} else {
		res.Message = bus.NewMessage(id, nil)
	}

	go func() {
		select {
		case <-b.ctx.Done():
			b.log.Warningf(`skip send %s: %v`, res.Message, b.ctx.Err())
		case b.resultChan <- res:
			b.log.Debugf(`sent %s`, res.Message)
		}
	}()
}

func (b *base) loop() {
	var res interface{}
	var err error
	b.log.Infof(`open %v`, b.config.Provider)
LOOP:
	for {
		select {
		case <-b.shutdownChan:
			if err = b.engine.shutdownFn(); err != nil {
				b.log.Errorf(`shutdown failed: %v`, err)
			}
			b.log.Info(`shutdown complete`)
			b.cancel()
			break LOOP
		case <-b.ctx.Done():
			break LOOP
		case op := <-b.opChan:
			b.log.Tracef(`accepted: %v`, op)
			switch op.op {
			case opResourceCreate:
				if res, err = b.engine.createFn(op.id, op.config, op.values); err != nil {
					b.log.Errorf(`create failed %v: %v`, op, err)
					continue LOOP
				}
				b.log.Infof(`created %v: %v`, op, res)
			case opResourceUpdate:
				if res, err = b.engine.updateFn(op.id, op.config); err != nil {
					b.log.Errorf(`update failed %v: %v`, op, err)
					continue LOOP
				}
				b.log.Infof(`updated %v: %v`, res, op)
			case opResourceDestroy:
				if err = b.engine.destroyFn(op.id); err != nil {
					b.log.Errorf(`destroy failed %s: %v`, op.id, err)
					continue LOOP
				}
				b.log.Infof(`destroyed: %s`, op.id)
			}
		}
	}
	b.log.Info(`closed`)
}
