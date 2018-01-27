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
	return &Lift{
		name:     name,
		consumer: consumer,
		catalog:  map[string]bus.Payload{},
	}
}

func (p *Lift) ConsumeMessage(message bus.Message) (err error) {
	if message.Topic() == "" {
		return p.consumeReset(message)
	}
	return p.consumeOne(message)
}

func (p *Lift) consumeReset(message bus.Message) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	var ingest map[string]map[string]string
	if err = message.Payload().Unmarshal(&ingest); err != nil {
		return err
	}
	catalog := map[string]bus.Payload{}
	for id, val := range ingest {
		catalog[id] = bus.NewPayload(val)
	}
	p.catalog = catalog
	return p.consumer.ConsumeMessage(p.makeMessage())
}

func (p *Lift) consumeOne(message bus.Message) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	topic := message.Topic()
	if current, exists := p.catalog[topic]; exists {
		if message.Payload().IsEmpty() {
			delete(p.catalog, topic)
			return p.consumer.ConsumeMessage(p.makeMessage())
		}
		if current.Hash() == message.Payload().Hash() {
			return nil
		}
	}
	if message.Payload().IsEmpty() {
		return nil
	}
	p.catalog[topic] = message.Payload()
	return p.consumer.ConsumeMessage(p.makeMessage())
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
	return bus.NewMessage(p.name, fields)
}
