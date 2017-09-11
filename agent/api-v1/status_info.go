package api_v1

import (
	"net/url"
	"context"
	"github.com/akaspin/supervisor"
	"github.com/akaspin/soil/agent"
	"sync"
)

type StatusInfoResponse map[string]*StatusInfoProducerData

type StatusInfoProducerData struct {
	Namespaces []string
	Active bool
	Data map[string]string
}

type statusInfoGetEndpoint struct {
	*supervisor.Control
	mu *sync.Mutex
	data StatusInfoResponse
	sources []agent.Source
}

func NewStatusInfoGetEndpoint(ctx context.Context, producers ...agent.Source) (e *statusInfoGetEndpoint)  {
	e = &statusInfoGetEndpoint{
		Control: supervisor.NewControl(ctx),
		mu: &sync.Mutex{},
		data: StatusInfoResponse{},
		sources: producers,
	}
	for _, producer := range producers {
		e.data[producer.Prefix()] = &StatusInfoProducerData{
			Active: false,
			Namespaces: producer.Namespaces(),
			Data: map[string]string{},
		}
	}
	return
}

func (e *statusInfoGetEndpoint) Open() (err error) {
	for _, producer := range e.sources {
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

func (e *statusInfoGetEndpoint) Sync(producer string, active bool, data map[string]string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if producer, ok := e.data[producer]; ok {
		producer.Data = data
		producer.Active = active
	}
	return
}

