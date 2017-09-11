package source

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/supervisor"
	"sync"
)

type BaseProducer struct {
	*supervisor.Control
	log *logx.Log

	prefix string

	mu        *sync.Mutex
	consumers []agent.SourceConsumer
	data      map[string]string
	active    bool
}

func NewBaseProducer(ctx context.Context, log *logx.Log, prefix string) (p *BaseProducer) {
	p = &BaseProducer{
		Control: supervisor.NewControl(ctx),
		log:     log.GetLog("producer", prefix),
		prefix:  prefix,
		mu:      &sync.Mutex{},
		data:    map[string]string{},
	}
	return
}

func (p *BaseProducer) Prefix() string {
	return p.prefix
}

func (p *BaseProducer) RegisterConsumer(name string, consumer agent.SourceConsumer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.consumers = append(p.consumers, consumer)
	p.log.Infof("%s consumer registered", name)
	p.notify()
}

func (p *BaseProducer) Store(active bool, data map[string]string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.data = data
	p.active = active
	p.notify()
}

func (p *BaseProducer) Put(active bool, data map[string]string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.active = active
	for k, v := range data {
		p.data[k] = v
	}
	p.notify()
}

func (p *BaseProducer) Delete(active bool, keys ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.active = active
	for _, k := range keys {
		delete(p.data, k)
	}
	p.notify()
}

func (p *BaseProducer) notify() {
	p.log.Debugf("syncing with %d consumers", len(p.consumers))
	for _, consumer := range p.consumers {
		consumer.Sync(p.prefix, p.active, p.data)
	}
}
