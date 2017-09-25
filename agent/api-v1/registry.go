package api_v1

import (
	"context"
	"encoding/json"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/public"
	"github.com/akaspin/soil/api"
	"github.com/akaspin/soil/api/api-v1-types"
	"github.com/akaspin/soil/manifest"
	"github.com/mitchellh/hashstructure"
	"net/url"
	"sync"
)

type registryGet struct {
	log *logx.Log
	data *sync.Map
}

func NewRegistryGet(log *logx.Log) (e *registryGet) {
	e = &registryGet{
		log: log.GetLog("api", "get", "/v1/registry"),
		data: &sync.Map{},
	}
	return
}

func (e *registryGet) Empty() interface{} {
	return nil
}

func (e *registryGet) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
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

func (e *registryGet) ConsumeRegistry(namespace string, payload manifest.Registry) {
	e.data.Store(namespace, payload)
}


type registryPut struct {
	log    *logx.Log
	setter public.Setter
}

func NewRegistryPut(log *logx.Log, setter public.Setter) (e *registryPut) {
	e = &registryPut{
		log:    log.GetLog("api", "put", "/v1/registry"),
		setter: setter,
	}
	return
}

func (e *registryPut) Empty() interface{} {
	return &api_v1_types.RegistryPutRequest{}
}

func (e *registryPut) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	pods, ok := v.(*api_v1_types.RegistryPutRequest)
	if !ok {
		err = api.NewError(500, "can't unmarshal pods")
		return
	}
	ingest := map[string]string{}
	marks := api_v1_types.RegistryPutResponse(map[string]uint64{})
	for _, pod := range *pods {
		data, marshalErr := json.Marshal(pod)
		if marshalErr != nil {
			e.log.Errorf("can't marshal pod: %v", err)
			continue
		}
		ingest["registry/"+pod.Name] = string(data)
		mark, _ := hashstructure.Hash(pod, nil)
		marks[pod.Name] = mark
	}
	e.setter.Set(ingest, false)
	res = marks
	return
}

type registryDelete struct {
	log *logx.Log
}

