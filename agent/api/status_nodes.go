package api

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/api/api-server"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/proto"
	"net/url"
	"sync"
)

func NewClusterNodesGet(log *logx.Log) (e *api_server.Endpoint) {
	return api_server.GET("/v1/status/nodes", &clusterNodesProcessor{
		log: log.GetLog("api", "get", "/v1/status/nodes"),
	})
}

type clusterNodesProcessor struct {
	log   *logx.Log
	mu    sync.Mutex
	nodes proto.NodesInfo
}

func (p *clusterNodesProcessor) Empty() interface{} {
	return nil
}

func (p *clusterNodesProcessor) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	res = p.nodes
	return
}

func (p *clusterNodesProcessor) ConsumeMessage(message bus.Message) {
	var v proto.NodesInfo
	if err := message.Payload().Unmarshal(&v); err != nil {
		p.log.Error(err)
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.nodes = v
}
