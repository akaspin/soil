package api

import (
	"context"
	"fmt"
	"github.com/akaspin/soil/agent/api/api-server"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/manifest"
	"github.com/kr/pretty"
	"net/url"
)

func NewRegistryPodPut(consumer bus.Consumer) (e *api_server.Endpoint) {
	return api_server.PUT("/v1/registry/pod", &registryPodPutProcessor{
		consumer: consumer,
	})
}

type registryPodPutProcessor struct {
	consumer bus.Consumer
}

func (p *registryPodPutProcessor) Empty() interface{} {
	return &manifest.Pod{
		Namespace: manifest.PublicNamespace,
	}
}

func (p *registryPodPutProcessor) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	pretty.Log(v)
	v1, ok := v.(*manifest.Pod)
	if !ok || v1 == nil {
		err = fmt.Errorf(`bad pod %v`, v)
		return
	}
	err = p.consumer.ConsumeMessage(bus.NewMessage(v1.Name, v1))
	return
}
