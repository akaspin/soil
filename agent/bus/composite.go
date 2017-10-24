package bus

import "sync"

// Composite pipe collects messages from declared sources
type CompositePipe struct {
	name       string
	downstream Consumer
	empty      Message

	mu      sync.Mutex
	sources map[string]Message
}

func NewCompositePipe(name string, downstream Consumer, source ...string) (p *CompositePipe) {
	p = &CompositePipe{
		name:       name,
		downstream: downstream,
		empty:      NewMessage(name, nil),
		sources:    map[string]Message{},
	}
	for _, m := range source {
		p.sources[m] = NewMessage(m, nil)
	}
	return
}

func (p *CompositePipe) ConsumeMessage(message Message) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, declared := p.sources[message.GetPrefix()]; !declared {
		// not in sources
		return
	}
	p.sources[message.GetPrefix()] = message
	if message.IsEmpty() {
		p.downstream.ConsumeMessage(p.empty)
		return
	}
	payload := map[string]string{}
	for prefix, msg := range p.sources {
		if msg.IsEmpty() {
			p.downstream.ConsumeMessage(p.empty)
			return
		}
		for k, v := range msg.GetPayload() {
			payload[prefix+"."+k] = v
		}
	}
	p.downstream.ConsumeMessage(NewMessage(p.name, payload))
	return
}
