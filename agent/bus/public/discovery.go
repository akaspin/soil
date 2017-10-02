package public

import (
	"encoding/json"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/proto"
	"strings"
)

type DiscoveryPipe struct {
	log *logx.Log
	*bus.SimplePipe
}

func NewDiscoveryPipe(log *logx.Log, consumer bus.MessageConsumer) (p *DiscoveryPipe) {
	p = &DiscoveryPipe{
		log: log.GetLog("pipe", "discovery"),
	}
	p.SimplePipe = bus.NewSimplePipe(p.process, consumer)
	return
}

func (p *DiscoveryPipe) process(message bus.Message) (res bus.Message) {
	data := map[string]string{}
	for _, v := range message.GetPayload() {
		var value proto.NodeResponse
		if err := json.NewDecoder(strings.NewReader(v)).Decode(&value); err != nil {
			p.log.Error(err)
		}
		data[value.Id] = value.Advertise
	}
	res = bus.NewMessage(message.GetProducer(), data)
	return
}
