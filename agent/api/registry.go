package api

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/api/api-server"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/manifest"
	"net/url"
	"sync"
)

func NewRegistryPodsGet() (e *api_server.Endpoint) {
	return api_server.GET("/v1/registry", &registryPodsGetProcessor{
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
	return api_server.PUT("/v1/registry", &registryPodsPutProcessor{
		log:      log.WithTags("put", "/v1/registry"),
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
		err = fmt.Errorf(`bad pods %v`, v)
		return
	}
	for _, pod := range *v1 {
		if consumeErr := p.consumer.ConsumeMessage(bus.NewMessage(pod.Name, pod)); consumeErr != nil {
			p.log.Error(err)
		}
	}
	return
}
