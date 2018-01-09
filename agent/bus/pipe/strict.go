package pipe

import (
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"sync"
)

// Strict pipe consumes flatmap messages only with predeclared IDs. If
// at least one of declared messages or consumed message is empty Strict pipe
// send empty message to downstream. If all messages are not empty Strict pipe
// combines all messages as <message-id>.<flatmap-key> = <flatmap-value>
// and sends result to downstream.
type StrictPipe struct {
	name       string
	log        *logx.Log
	downstream bus.Consumer
	empty      bus.Message
	mu         sync.Mutex
	declared   map[string]bus.Message
}

func NewStrict(name string, log *logx.Log, downstream bus.Consumer, declared ...string) (p *StrictPipe) {
	p = &StrictPipe{
		name:       name,
		log:        log.GetLog("pipe", "strict", name),
		downstream: downstream,
		empty:      bus.NewMessage(name, nil),
		declared:   map[string]bus.Message{},
	}
	for _, m := range declared {
		p.declared[m] = bus.NewMessage(m, nil)
	}
	return
}

func (p *StrictPipe) ConsumeMessage(message bus.Message) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, declared := p.declared[message.GetID()]; !declared {
		return
	}
	p.declared[message.GetID()] = message
	if message.Payload().IsEmpty() {
		p.downstream.ConsumeMessage(p.empty)
		return
	}
	payload := map[string]string{}
	for prefix, msg := range p.declared {
		if msg.Payload().IsEmpty() {
			p.downstream.ConsumeMessage(p.empty)
			return
		}
		var chunk map[string]string

		if mErr := msg.Payload().Unmarshal(&chunk); mErr != nil {
			p.log.Error(mErr)
			continue
		}
		for k, v := range chunk {
			payload[prefix+"."+k] = v
		}
	}
	err = p.downstream.ConsumeMessage(bus.NewMessage(p.name, payload))
	return
}
