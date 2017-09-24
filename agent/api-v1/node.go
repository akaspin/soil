package api_v1

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/metadata"
	"net/url"
	"sync"
)

type statusNode struct {
	log  *logx.Log
	data *sync.Map
}

func NewStatusNode(log *logx.Log) (e *statusNode) {
	e = &statusNode{
		log:  log.GetLog("api", "v1", "status/node"),
		data: &sync.Map{},
	}
	return
}

func (n *statusNode) Empty() interface{} {
	return nil
}

func (n *statusNode) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	res1 := map[string]map[string]string{}
	n.data.Range(func(key, value interface{}) bool {
		res1[key.(string)] = value.(map[string]string)
		return true
	})
	res = res1
	return
}

func (n *statusNode) ConsumeMessage(message metadata.Message) {
	if !message.Clean {
		return
	}
	n.data.Store(message.Prefix, message.Data)
	n.log.Debugf("stored: %s=%v", message.Prefix, message.Data)
}
