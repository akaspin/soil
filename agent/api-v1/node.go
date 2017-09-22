package api_v1

import (
	"context"
	"github.com/akaspin/soil/agent/metadata"
	"github.com/akaspin/soil/api/api-v1-types"
	"net/url"
	"sync"
	"encoding/json"
	"strings"
	"github.com/akaspin/logx"
)

type ClusterNodesGET struct {
	log *logx.Log
	mu *sync.RWMutex
	data api_v1_types.NodesResponse
}

func NewClusterNodesGET(log *logx.Log) (e *ClusterNodesGET) {
	e = &ClusterNodesGET{
		log: log.GetLog("api", "public/nodes"),
		mu: &sync.RWMutex{},
	}
	return
}

func (e *ClusterNodesGET) Empty() interface{} {
	return nil
}

func (e *ClusterNodesGET) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	res = e.data
	return
}

func (e *ClusterNodesGET) Sync(message metadata.Message) {
	if !message.Clean {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.data = e.data[:]
	for _, v := range message.Data {
		var val api_v1_types.NodeResponse
		if err := json.NewDecoder(strings.NewReader(v)).Decode(&val); err != nil {
			e.log.Error(err)
			continue
		}
		e.data = append(e.data, val)
	}

}



