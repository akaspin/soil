package api

import (
	"context"
	"net/url"
)

// simple wrapper for non-processing handlers
type Wrapper struct {
	fn func() (err error)
}

func NewWrapper(fn func() (err error)) (e *Wrapper) {
	return &Wrapper{
		fn: fn,
	}
}

func (e *Wrapper) Empty() interface{} {
	return nil
}

func (e *Wrapper) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	if err = e.fn(); err != nil {
		return nil, err
	}
	return "ok", nil
}
