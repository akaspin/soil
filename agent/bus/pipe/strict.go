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
	topics     map[string]bus.Message
}

func NewStrict(name string, log *logx.Log, downstream bus.Consumer, declared ...string) (p *StrictPipe) {
	p = &StrictPipe{
		name:       name,
		log:        log.GetLog("pipe", "strict", name),
		downstream: downstream,
		empty:      bus.NewMessage(name, nil),
		topics:     map[string]bus.Message{},
	}
	for _, m := range declared {
		p.topics[m] = bus.NewMessage(m, nil)
	}
	return p
}

func (p *StrictPipe) ConsumeMessage(message bus.Message) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, declared := p.topics[message.Topic()]; !declared {
		return nil
	}
	p.topics[message.Topic()] = message
	if message.Payload().IsEmpty() {
		return p.downstream.ConsumeMessage(p.empty)
	}
	payload := map[string]string{}
	for prefix, msg := range p.topics {
		if msg.Payload().IsEmpty() {
			return p.downstream.ConsumeMessage(p.empty)
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
	return p.downstream.ConsumeMessage(bus.NewMessage(p.name, payload))
}
