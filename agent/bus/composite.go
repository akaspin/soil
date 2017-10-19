package bus

import "sync"

// Composite pipe collects messages from declared sources
type CompositePipe struct {
	name     string
	consumer MessageConsumer

	empty Message

	mu      sync.Mutex
	sources map[string]Message
}

func NewCompositePipe(name string, consumer MessageConsumer, source ...string) (p *CompositePipe) {
	p = &CompositePipe{
		name:     name,
		consumer: consumer,
		empty:    NewMessage(name, nil),
		sources:  map[string]Message{},
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
		p.consumer.ConsumeMessage(p.empty)
		return
	}
	payload := map[string]string{}
	for prefix, msg := range p.sources {
		if msg.IsEmpty() {
			p.consumer.ConsumeMessage(p.empty)
			return
		}
		for k, v := range msg.GetPayload() {
			payload[prefix+"."+k] = v
		}
	}
	p.consumer.ConsumeMessage(NewMessage(p.name, payload))
	return
}
