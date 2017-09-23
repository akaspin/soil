package api

import (
	"encoding/json"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/metadata"
	"github.com/akaspin/soil/api/api-v1-types"
	"strings"
)

type DiscoveryPipe struct {
	log *logx.Log
	*metadata.SimplePipe
}

func NewDiscoveryPipe(log *logx.Log, router *Router) (p *DiscoveryPipe) {
	p = &DiscoveryPipe{
		log: log.GetLog("pipe", "discovery"),
	}
	p.SimplePipe = metadata.NewSimplePipe(p.process, router.Sync)
	return
}

func (p *DiscoveryPipe) process(message metadata.Message) (res metadata.Message) {
	if !message.Clean {
		return
	}
	data := map[string]string{}
	for _, v := range message.Data {
		var value api_v1_types.NodeResponse
		if err := json.NewDecoder(strings.NewReader(v)).Decode(&value); err != nil {
			p.log.Error(err)
		}
		data[value.Id] = value.Advertise
	}
	res = message
	res.Data = data
	return
}
