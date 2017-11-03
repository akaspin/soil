package bus

import (
	"github.com/akaspin/logx"
	"sync"
)

// Composite pipe
type CompositePipe struct {
	name       string
	log        *logx.Log
	downstream Consumer
	empty      Message
	mu         sync.Mutex
	declared   map[string]Message
}

func NewCompositePipe(name string, log *logx.Log, downstream Consumer, declared ...string) (p *CompositePipe) {
	p = &CompositePipe{
		name:       name,
		log:        log.GetLog("pipe", "composite", name),
		downstream: downstream,
		empty:      NewMessage(name, nil),
		declared:   map[string]Message{},
	}
	for _, m := range declared {
		p.declared[m] = NewMessage(m, nil)
	}
	return
}

func (p *CompositePipe) ConsumeMessage(message Message) {
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
		err := msg.Payload().Unmarshal(&chunk)
		if err != nil {
			p.log.Error(err)
			continue
		}
		for k, v := range chunk {
			payload[prefix+"."+k] = v
		}
	}
	p.downstream.ConsumeMessage(NewMessage(p.name, payload))
	return
}
