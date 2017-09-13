package api_v1

import (
	"context"
	"net/url"
)

type DrainResponse struct {
	AgentId string
	Drain   bool
}

type drainGetEndpoint struct {
	agentId string
	fn      func() bool
}

func NewDrainGetEndpoint(agentId string, fn func() bool) (e *drainGetEndpoint) {
	e = &drainGetEndpoint{
		agentId: agentId,
		fn:      fn,
	}
	return
}

func (e *drainGetEndpoint) Empty() interface{} {
	return nil
}

func (e *drainGetEndpoint) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	res = DrainResponse{
		AgentId: e.agentId,
		Drain:   e.fn(),
	}
	return
}

type drainPutEndpoint struct {
	drainFn func(state bool)
}

func NewDrainPutEndpoint(drainFn func(state bool)) (e *drainPutEndpoint) {
	e = &drainPutEndpoint{
		drainFn: drainFn,
	}
	return
}

func (e *drainPutEndpoint) Empty() interface{} {
	return nil
}

func (e *drainPutEndpoint) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	e.drainFn(true)
	return
}

type drainDeleteEndpoint struct {
	drainFn func(state bool)
}

func NewDrainDeleteEndpoint(drainFn func(state bool)) (e *drainDeleteEndpoint) {
	e = &drainDeleteEndpoint{
		drainFn: drainFn,
	}
	return
}

func (e *drainDeleteEndpoint) Empty() interface{} {
	return nil
}

func (e *drainDeleteEndpoint) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	e.drainFn(false)
	return
}
