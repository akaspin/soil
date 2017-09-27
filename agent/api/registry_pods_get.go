package api

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/api/api-server"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/soil/proto"
	"net/http"
	"net/url"
	"sync"
)

func NewRegistryPodsGet(log *logx.Log) (e *api_server.Endpoint) {
	return api_server.NewEndpoint(http.MethodGet, proto.V1RegistryPods,
		&registryPodsGetProcessor{
			log:  log.GetLog("api", "get", proto.V1RegistryPods),
			data: &sync.Map{},
		})
}

type registryPodsGetProcessor struct {
	log  *logx.Log
	data *sync.Map
}

func (e *registryPodsGetProcessor) Empty() interface{} {
	return nil
}

func (e *registryPodsGetProcessor) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	r1 := map[string]manifest.Registry{}
	e.data.Range(func(key, value interface{}) bool {
		r2, ok := value.(manifest.Registry)
		if !ok {
			e.log.Errorf("can't load registry (%T)%v", value, value)
			return true
		}
		r1[key.(string)] = r2
		return true
	})
	res = r1
	return
}

func (e *registryPodsGetProcessor) ConsumeRegistry(namespace string, payload manifest.Registry) {
	e.data.Store(namespace, payload)
}
