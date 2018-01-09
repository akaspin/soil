package pipe

import (
	"github.com/akaspin/soil/agent/bus"
	"sync"
)

// Lift pipe holds set catalog of consumed map[string]string messages and
// propagates them to downstream as name:map[message-id+k]=v. To reset catalog
// pipe send map[id]map[k]string message with empty message id.
type Lift struct {
	name     string
	consumer bus.Consumer
	mu       sync.Mutex
	catalog  map[string]bus.Payload
}

func NewLift(name string, consumer bus.Consumer) (p *Lift) {
	p = &Lift{
		name:     name,
		consumer: consumer,
		catalog:  map[string]bus.Payload{},
	}
	return
}

func (p *Lift) ConsumeMessage(message bus.Message) (err error) {
	if message.Topic() == "" {
		err = p.consumeReset(message)
		return
	}
	err = p.consumeOne(message)
	return
}

func (p *Lift) consumeReset(message bus.Message) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	var ingest map[string]map[string]string
	if err = message.Payload().Unmarshal(&ingest); err != nil {
		return
	}
	catalog := map[string]bus.Payload{}
	for id, val := range ingest {
		catalog[id] = bus.NewPayload(val)
	}
	p.catalog = catalog
	err = p.consumer.ConsumeMessage(p.makeMessage())
	return
}

func (p *Lift) consumeOne(message bus.Message) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	topic := message.Topic()
	if current, exists := p.catalog[topic]; exists {
		if message.Payload().IsEmpty() {
			delete(p.catalog, topic)
			err = p.consumer.ConsumeMessage(p.makeMessage())
			return
		}
		if current.Hash() == message.Payload().Hash() {
			return
		}
	}
	if message.Payload().IsEmpty() {
		return
	}
	p.catalog[topic] = message.Payload()
	err = p.consumer.ConsumeMessage(p.makeMessage())
	return
}

func (p *Lift) makeMessage() (res bus.Message) {
	fields := map[string]string{}
	for root, payload := range p.catalog {
		var data map[string]string
		payload.Unmarshal(&data)
		for k, v := range data {
			fields[root+"."+k] = v
		}
	}
	res = bus.NewMessage(p.name, fields)
	return
}
