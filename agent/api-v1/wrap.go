package api_v1

import (
	"context"
	"net/url"
)

// simple wrapper for non-processing handlers
type Wrapper struct {
	fn func() (err error)
}

func NewWrapper(fn func() (err error)) (e *Wrapper) {
	e = &Wrapper{
		fn: fn,
	}
	return
}

func (e *Wrapper) Empty() interface{} {
	return nil
}

func (e *Wrapper) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	err = e.fn()
	if err = e.fn(); err == nil {
		res = "ok"
	}
	return
}
