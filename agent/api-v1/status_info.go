package api_v1

import (
	"context"
	"github.com/akaspin/soil/agent/metadata"
	"github.com/akaspin/supervisor"
	"net/url"
	"sync"
)

type StatusInfoResponse map[string]metadata.Message

type statusInfoGetEndpoint struct {
	*supervisor.Control
	mu        *sync.Mutex
	data      StatusInfoResponse
	producers []metadata.Producer
}

func NewStatusInfoGetEndpoint(ctx context.Context, producers ...metadata.Producer) (e *statusInfoGetEndpoint) {
	e = &statusInfoGetEndpoint{
		Control:   supervisor.NewControl(ctx),
		mu:        &sync.Mutex{},
		data:      StatusInfoResponse{},
		producers: producers,
	}
	return
}

func (e *statusInfoGetEndpoint) Open() (err error) {
	for _, producer := range e.producers {
		producer.RegisterConsumer("api-status-info", e)
	}
	err = e.Control.Open()
	return
}

func (e *statusInfoGetEndpoint) Empty() interface{} {
	return nil
}

func (e *statusInfoGetEndpoint) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	res = e.data
	return
}

func (e *statusInfoGetEndpoint) Sync(message metadata.Message) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.data[message.Prefix] = message
}
