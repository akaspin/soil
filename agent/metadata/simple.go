package metadata

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/supervisor"
	"sync"
)

type SimpleProducer struct {
	*supervisor.Control
	log *logx.Log

	prefix string

	mu        *sync.Mutex
	consumers []func(message Message)
	data      map[string]string
	active    bool
}

func NewSimpleProducer(ctx context.Context, log *logx.Log, prefix string, consumers ...func(message Message)) (p *SimpleProducer) {
	p = &SimpleProducer{
		Control: supervisor.NewControl(ctx),
		log:     log.GetLog("producer", prefix),
		prefix:  prefix,
		mu:      &sync.Mutex{},
		consumers: consumers,
		data:    map[string]string{},
	}
	return
}

func (p *SimpleProducer) Open() (err error) {
	p.log.Debug("open")
	err = p.Control.Open()
	return
}

func (p *SimpleProducer) Close() error {
	p.log.Debug("close")
	return p.Control.Close()
}

func (p *SimpleProducer) Prefix() string {
	return p.prefix
}

func (p *SimpleProducer) RegisterConsumer(name string, fn func(message Message)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.consumers = append(p.consumers, fn)
	p.log.Infof("registered consumer: %s", name)
	if p.active {
		fn(Message{
			Prefix: p.prefix,
			Clean:  p.active,
			Data:   p.data,
		})
		p.log.Debugf("consumer %s notified: %v", name, p.data)
	}
}

func (p *SimpleProducer) Replace(data map[string]string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.data = data
	p.active = data != nil
	p.notifyAll()
}

func (p *SimpleProducer) Set(active bool, data map[string]string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.active = active
	for k, v := range data {
		p.data[k] = v
	}
	p.notifyAll()
}

func (p *SimpleProducer) Delete(active bool, keys ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.active = active
	for _, k := range keys {
		delete(p.data, k)
	}
	p.notifyAll()
}

func (p *SimpleProducer) notifyAll() {
	p.log.Tracef("syncing with %d consumers", len(p.consumers))
	for _, consumer := range p.consumers {
		consumer(Message{
			Prefix: p.prefix,
			Clean:  p.active,
			Data:   p.data,
		})
	}
}
