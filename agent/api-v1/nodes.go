package api_v1

import (
	"context"
	"encoding/json"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/metadata"
	"github.com/akaspin/soil/api/api-v1-types"
	"net/url"
	"strings"
	"sync"
)

type StatusNodes struct {
	log  *logx.Log
	mu   *sync.RWMutex
	data api_v1_types.NodesResponse
}

func NewStatusNodes(log *logx.Log) (e *StatusNodes) {
	e = &StatusNodes{
		log: log.GetLog("api", "status/nodes"),
		mu:  &sync.RWMutex{},
	}
	return
}

func (e *StatusNodes) Empty() interface{} {
	return nil
}

func (e *StatusNodes) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	res = e.data
	return
}

func (e *StatusNodes) ConsumeMessage(message metadata.Message) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.data = e.data[:0]
	for _, v := range message.GetPayload() {
		var val api_v1_types.NodeResponse
		if err := json.NewDecoder(strings.NewReader(v)).Decode(&val); err != nil {
			e.log.Error(err)
			continue
		}
		e.data = append(e.data, val)
	}
}
