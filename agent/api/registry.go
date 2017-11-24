package api

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/api/api-server"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/manifest"
	"net/http"
	"net/url"
	"sync"
)

const (
	V1Registry = "/v1/registry"
)

func NewRegistryPodsGet() (e *api_server.Endpoint) {
	return api_server.GET(V1Registry, &registryPodsGetProcessor{
		pods: manifest.Registry{},
	})
}

type registryPodsGetProcessor struct {
	mu   sync.Mutex
	pods manifest.Registry
}

func (p *registryPodsGetProcessor) Empty() interface{} {
	return nil
}

func (p *registryPodsGetProcessor) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	res = p.pods
	return
}

func (p *registryPodsGetProcessor) ConsumeMessage(message bus.Message) (err error) {
	var v manifest.Registry
	if err = message.Payload().Unmarshal(&v); err != nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pods = v
	return
}

func NewRegistryPodsPut(log *logx.Log, consumer bus.Consumer) (e *api_server.Endpoint) {
	return api_server.PUT(V1Registry, &registryPodsPutProcessor{
		log:      log.GetLog("api", "put", V1Registry),
		consumer: consumer,
	})
}

type registryPodsPutProcessor struct {
	log      *logx.Log
	consumer bus.Consumer
}

func (p *registryPodsPutProcessor) Empty() interface{} {
	return &manifest.Registry{}
}

func (p *registryPodsPutProcessor) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	v1, ok := v.(*manifest.Registry)
	if !ok || v1 == nil || len(*v1) == 0 {
		err = api_server.NewError(http.StatusBadRequest, fmt.Sprintf("bad pods: %v", v))
		return
	}
	for _, pod := range *v1 {
		if consumeErr := p.consumer.ConsumeMessage(bus.NewMessage(pod.Name, pod)); consumeErr != nil {
			p.log.Error(err)
		}
	}
	return
}

func NewRegistryPodsDelete(log *logx.Log, consumer bus.Consumer) (e *api_server.Endpoint) {
	return api_server.DELETE(V1Registry, &registryPodsDeleteProcessor{
		log:      log.GetLog("api", "delete", V1Registry),
		consumer: consumer,
	})
}

type registryPodsDeleteProcessor struct {
	log      *logx.Log
	consumer bus.Consumer
}

func (p *registryPodsDeleteProcessor) Empty() interface{} {
	return nil
}

func (p *registryPodsDeleteProcessor) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	pods, ok := u.Query()["pod"]
	if !ok || pods == nil || len(pods) == 0 {
		err = api_server.NewError(http.StatusBadRequest, fmt.Sprintf("bad pod query: %s", u.RawQuery))
		return
	}
	for _, pod := range pods {
		if consumeErr := p.consumer.ConsumeMessage(bus.NewMessage(pod, nil)); consumeErr != nil {
			p.log.Error(err)
		}
	}
	return
}
