package api_v1

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/api-v1/api-server"
	"github.com/akaspin/soil/agent/bus"
	"net/url"
	"sync"
)

func NewStatusNodeGet(log *logx.Log) (e *api_server.Endpoint) {
	return api_server.GET("/v1/status/node", &statusNodeProcessor{
		log:  log.GetLog("api", "get", "/v1/status/node"),
		data: &sync.Map{},
	})
}

type statusNodeProcessor struct {
	log  *logx.Log
	data *sync.Map
}

func (n *statusNodeProcessor) Empty() interface{} {
	return nil
}

func (n *statusNodeProcessor) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	res1 := map[string]map[string]string{}
	n.data.Range(func(key, value interface{}) bool {
		res1[key.(string)] = value.(map[string]string)
		return true
	})
	res = res1
	return
}

func (n *statusNodeProcessor) ConsumeMessage(message bus.Message) {
	n.data.Store(message.GetPrefix(), message.GetPayload())
	n.log.Debugf("stored: %v", message)
}
