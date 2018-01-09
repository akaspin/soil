package bus

import (
	"github.com/akaspin/logx"
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
	downstream Consumer
	empty      Message
	mu         sync.Mutex
	declared   map[string]Message
}

func NewStrictPipe(name string, log *logx.Log, downstream Consumer, declared ...string) (p *StrictPipe) {
	p = &StrictPipe{
		name:       name,
		log:        log.GetLog("pipe", "strict", name),
		downstream: downstream,
		empty:      NewMessage(name, nil),
		declared:   map[string]Message{},
	}
	for _, m := range declared {
		p.declared[m] = NewMessage(m, nil)
	}
	return
}

func (p *StrictPipe) ConsumeMessage(message Message) (err error) {
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
	err = p.downstream.ConsumeMessage(NewMessage(p.name, payload))
	return
}
