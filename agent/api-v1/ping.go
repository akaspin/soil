package api_v1

import (
	"context"
	"net/url"
)

// PingEndpoint serves "/v1/status/ping" requests
type PingEndpoint struct{}

func NewPingEndpoint() (e *PingEndpoint) {
	e = &PingEndpoint{}
	return
}

func (e *PingEndpoint) Empty() interface{} {
	return nil
}

func (e *PingEndpoint) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	res = "ok"
	return
}

type PingResponse struct {
	Id string
}
