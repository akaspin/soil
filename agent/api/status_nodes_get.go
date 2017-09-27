package api

import (
	"context"
	"encoding/json"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/api/api-server"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/proto"
	"net/url"
	"strings"
	"sync"
)

func NewStatusNodesGet(log *logx.Log) (e *api_server.Endpoint) {
	return api_server.GET("/v1/status/nodes", &statusNodesProcessor{
		log: log.GetLog("api", "get", "/v1/status/nodes"),
		mu:  &sync.RWMutex{},
	})
}

type statusNodesProcessor struct {
	log  *logx.Log
	mu   *sync.RWMutex
	data proto.NodesResponse
}

func (e *statusNodesProcessor) Empty() interface{} {
	return nil
}

func (e *statusNodesProcessor) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	res = e.data
	return
}

func (e *statusNodesProcessor) ConsumeMessage(message bus.Message) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.data = e.data[:0]
	for _, v := range message.GetPayload() {
		var val proto.NodeResponse
		if err := json.NewDecoder(strings.NewReader(v)).Decode(&val); err != nil {
			e.log.Error(err)
			continue
		}
		e.data = append(e.data, val)
	}
}
