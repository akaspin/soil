package api

import (
	"context"
	"github.com/akaspin/soil/agent/api/api-server"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/proto"
	"net/http"
	"net/url"
)

func NewRegistryPodsDelete(deleter bus.Deleter) (e *api_server.Endpoint) {
	return api_server.NewEndpoint(http.MethodDelete, proto.V1RegistryPods, &registryPodsDeleteProcessor{
		deleter: deleter,
	})
}

type registryPodsDeleteProcessor struct {
	deleter bus.Deleter
}

func (p *registryPodsDeleteProcessor) Empty() interface{} {
	return &proto.RegistryPodsDeleteRequest{}
}

func (p *registryPodsDeleteProcessor) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	data, ok := v.(*proto.RegistryPodsDeleteRequest)
	if !ok {
		err = api_server.ErrorBadRequestData
		return
	}
	p.deleter.Delete(*data...)
	return
}
