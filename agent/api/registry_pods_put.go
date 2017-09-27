package api

import (
	"context"
	"encoding/json"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/api/api-server"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/proto"
	"net/http"
	"net/url"
)

func NewRegistryPodsPut(log *logx.Log, setter bus.Setter) (e *api_server.Endpoint) {
	e = api_server.NewEndpoint(http.MethodPut, proto.V1RegistryPods,
		&registryPodsPutProcessor{
			log:    log.GetLog("api", "put", proto.V1RegistryPods),
			setter: setter,
		})
	return
}

type registryPodsPutProcessor struct {
	log    *logx.Log
	setter bus.Setter
}

func (e *registryPodsPutProcessor) Empty() interface{} {
	return &proto.RegistryPodsPutRequest{}
}

func (e *registryPodsPutProcessor) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	pods, ok := v.(*proto.RegistryPodsPutRequest)
	if !ok {
		err = api_server.NewError(500, "can't unmarshal pods")
		return
	}
	ingest := map[string]string{}
	for _, pod := range *pods {
		data, marshalErr := json.Marshal(pod)
		if marshalErr != nil {
			e.log.Errorf("can't marshal pod: %v", err)
			continue
		}
		ingest[pod.Name] = string(data)
	}
	e.setter.Set(ingest)
	return
}
