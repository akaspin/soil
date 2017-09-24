package api_v1

import (
	"context"
	"encoding/json"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/public/kv"
	"github.com/akaspin/soil/api"
	"github.com/akaspin/soil/api/api-v1-types"
	"github.com/mitchellh/hashstructure"
	"net/url"
	"sync"
	"github.com/akaspin/soil/agent/metadata"
)

type registryGet struct {
	mu *sync.Mutex
	data api_v1_types.RegistryGetResponse
}

func NewRegistryGet() (e *registryGet) {
	e = &registryGet{
		mu: &sync.Mutex{},
	}
	return
}

func (e *registryGet) Empty() interface{} {
	return nil
}

func (e *registryGet) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	return
}

func (e *registryGet) Sync(message metadata.Message) {


	return
}

type registryPut struct {
	log *logx.Log
	setter kv.Setter
}

func NewRegistryPut(log *logx.Log, setter kv.Setter) (e *registryPut) {
	e = &registryPut{
		log: log.GetLog("api", "put /v1/pods"),
		setter: setter,
	}
	return
}

func (e *registryPut) Empty() interface{} {
	return &api_v1_types.RegistrySubmitRequest{}
}

func (e *registryPut) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	pods, ok := v.(*api_v1_types.RegistrySubmitRequest)
	if !ok {
		err = api.NewError(500, "can't unmarshal pods")
		return
	}
	ingest := map[string]string{}
	marks := api_v1_types.RegistrySubmitResponse(map[string]uint64{})
	for _, pod := range *pods {
		data, marshalErr := json.Marshal(pod)
		if marshalErr != nil {
			e.log.Errorf("can't marshal pod: %v", err)
			continue
		}
		ingest["registry/" + pod.Name] = string(data)
		mark, _ := hashstructure.Hash(pod, nil)
		marks[pod.Name] = mark
	}
	e.setter.Set(ingest, false)
	res = marks
	return
}

