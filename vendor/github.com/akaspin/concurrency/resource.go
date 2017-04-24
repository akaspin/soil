package concurrency

import (
	"io"
	"time"
	"github.com/akaspin/supervisor"
	"context"
	"errors"
)

var (
	PoolClosedError = errors.New("pool closed")
	PoolIsFullError = errors.New("pool is full")
)

type Resource interface {
	io.Closer
}

type Factory func() (Resource, error)

type resourceWrapper struct {
	resource Resource
	timeUsed time.Time
}


type ResourcePool struct {
	control     *supervisor.Control
	resourcesCh chan resourceWrapper
	config      Config
	factory     Factory
}

func NewResourcePool(ctx context.Context, config Config, factory Factory) (p *ResourcePool) {
	p = &ResourcePool{
		control: supervisor.NewControl(ctx),
		resourcesCh: make(chan resourceWrapper, config.Capacity),
		config: config,
		factory: factory,
	}
	return
}

func (p *ResourcePool) Open() (err error) {
	if err = p.control.Open(); err != nil {
		return
	}
	for i := 0; i < p.config.Capacity; i++ {
		p.resourcesCh<- resourceWrapper{}
	}
	return
}

func (p *ResourcePool) Close() (err error) {
	LOOP:
	for {
		select {
		case r := <-p.resourcesCh:
			if r.resource != nil {
				r.resource.Close()
			}
		default:
			break LOOP
		}
	}
	p.control.Close()
	return
}

func (p *ResourcePool) Wait() (err error) {
	err = p.control.Wait()
	return
}

func (p *ResourcePool) Get(ctx context.Context) (r Resource, err error) {
	select {
	case <-p.control.Ctx().Done():
		err = PoolClosedError
		return
	case <-ctx.Done():
		err = ctx.Err()
	default:
	}

	var wrapper resourceWrapper
	var ok bool
	select {
	case wrapper, ok = <-p.resourcesCh:
	case <-p.control.Ctx().Done():
		err = PoolClosedError
		return
	case <-ctx.Done():
		err = context.Canceled
	}
	if !ok {
		err = ctx.Err()
		return
	}

	if wrapper.resource != nil && p.config.IdleTimeout > 0 &&
			wrapper.timeUsed.Add(p.config.IdleTimeout).Sub(time.Now()) < 0 {
		wrapper.resource.Close()
		wrapper.resource = nil
	}
	if wrapper.resource == nil {
		wrapper.resource, err = p.factory()
		if err != nil {
			p.resourcesCh<- resourceWrapper{}
			return
		}
	}
	r = wrapper.resource
	return
}

func (p *ResourcePool) Put(r Resource) (err error) {
	select {
	case <-p.control.Ctx().Done():
		if r != nil {
			r.Close()
		}
		err = PoolClosedError
		return
	default:
	}

	var wrapper resourceWrapper
	if r != nil {
		wrapper = resourceWrapper{
			resource: r,
			timeUsed: time.Now(),
		}
	}
	select {
	case p.resourcesCh<- wrapper:
	default:
		err = PoolIsFullError
	}
	return
}

