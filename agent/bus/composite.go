package bus

import "sync"

// Composite pipe
type CompositePipe struct {
	name       string
	downstream Consumer
	empty      Message
	mu         sync.Mutex
	declared   map[string]Message
}

func NewCompositePipe(name string, downstream Consumer, declared ...string) (p *CompositePipe) {
	p = &CompositePipe{
		name:       name,
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
	if message.IsEmpty() {
		p.downstream.ConsumeMessage(p.empty)
		return
	}
	payload := map[string]string{}
	for prefix, msg := range p.declared {
		if msg.IsEmpty() {
			p.downstream.ConsumeMessage(p.empty)
			return
		}
		for k, v := range msg.GetPayloadMap() {
			payload[prefix+"."+k] = v
		}
	}
	p.downstream.ConsumeMessage(NewMessage(p.name, payload))
	return
}
