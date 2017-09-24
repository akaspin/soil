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

	prefix    string
	consumers []Consumer

	mu   *sync.Mutex
	data map[string]string
}

func NewSimpleProducer(ctx context.Context, log *logx.Log, prefix string, consumers ...Consumer) (p *SimpleProducer) {
	p = &SimpleProducer{
		Control:   supervisor.NewControl(ctx),
		log:       log.GetLog("producer", prefix),
		prefix:    prefix,
		consumers: consumers,
		mu:        &sync.Mutex{},
		data:      map[string]string{},
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

func (p *SimpleProducer) Replace(data map[string]string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.data = data
	p.notifyAll()
}

func (p *SimpleProducer) Set(data map[string]string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for k, v := range data {
		p.data[k] = v
	}
	p.notifyAll()
}

func (p *SimpleProducer) notifyAll() {
	p.log.Tracef("syncing with %d consumers", len(p.consumers))
	msg := NewMessage(p.prefix, p.data)
	for _, consumer := range p.consumers {
		consumer.ConsumeMessage(msg)
	}
}
