package bus

import (
	"sync"
)

// Catalog pipe holds set catalog of consumed map[string]string messages and propagates them to downstream as name:map[message-id+k]=v. To reset catalog pipe send map[id]map[k]string message with empty message id
type CatalogPipe struct {
	name     string
	consumer Consumer
	mu       sync.Mutex
	catalog  map[string]Payload
}

func NewCatalogPipe(name string, consumer Consumer) (p *CatalogPipe) {
	p = &CatalogPipe{
		name:     name,
		consumer: consumer,
		catalog:  map[string]Payload{},
	}
	return
}

func (p *CatalogPipe) ConsumeMessage(message Message) (err error) {
	if message.GetID() == "" {
		err = p.consumeReset(message)
		return
	}
	err = p.consumeOne(message)
	return
}

func (p *CatalogPipe) consumeReset(message Message) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	var ingest map[string]map[string]string
	if err = message.Payload().Unmarshal(&ingest); err != nil {
		return
	}
	catalog := map[string]Payload{}
	for id, val := range ingest {
		catalog[id] = NewPayload(val)
	}
	p.catalog = catalog
	err = p.consumer.ConsumeMessage(p.makeMessage())
	return
}

func (p *CatalogPipe) consumeOne(message Message) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	id := message.GetID()
	if current, exists := p.catalog[id]; exists {
		if message.Payload().IsEmpty() {
			delete(p.catalog, id)
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
	p.catalog[id] = message.Payload()
	err = p.consumer.ConsumeMessage(p.makeMessage())
	return
}

func (p *CatalogPipe) makeMessage() (res Message) {
	fields := map[string]string{}
	for root, payload := range p.catalog {
		var data map[string]string
		payload.Unmarshal(&data)
		for k, v := range data {
			fields[root+"."+k] = v
		}
	}
	res = NewMessage(p.name, fields)
	return
}
