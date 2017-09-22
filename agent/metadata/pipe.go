package metadata

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/supervisor"
	"sync"
)

type SimplePipe struct {
	consumers []func(message Message)
	fn func(message Message) Message
}

func NewSimplePipe(fn func(message Message) Message, consumers ...func(message Message)) (p *SimplePipe) {
	p = &SimplePipe{
		fn: fn,
		consumers: consumers,
	}
	return
}

func (p *SimplePipe) Sync(message Message) {
	res := message
	if p.fn != nil {
		res = p.fn(res)
	}
	for _, consumer := range p.consumers {
		consumer(res)
	}
}

type Pipe struct {
	*supervisor.Control
	log      *logx.Log
	prefix   string
	producer Producer
	fn       func(message Message) Message

	mu        *sync.Mutex
	cache     Message
	consumers []func(message Message)
}

func NewPipe(ctx context.Context, log *logx.Log, prefix string, producer Producer, fn func(message Message) Message, consumers ...func(message Message)) (p *Pipe) {
	p = &Pipe{
		Control:  supervisor.NewControl(ctx),
		log:      log.GetLog("pipe", prefix),
		prefix:   prefix,
		producer: producer,
		fn:       fn,
		mu:       &sync.Mutex{},
		consumers: consumers,
	}
	return
}

func (p *Pipe) Open() (err error) {
	go p.producer.RegisterConsumer(p.prefix, p.Sync)
	err = p.Control.Open()
	return
}

func (p *Pipe) Prefix() string {
	return p.prefix
}

func (p *Pipe) RegisterConsumer(name string, consumer func(message Message)) {
	go func() {
		p.mu.Lock()
		cache := p.cache
		p.consumers = append(p.consumers, consumer)
		p.mu.Unlock()
		p.log.Debugf("registered consumer: %s", name)
		if cache.Clean {
			consumer(cache)
		}
	}()
}

func (p *Pipe) Sync(message Message) {
	p.log.Debugf("accepted message %v", message)
	res := message
	if p.fn != nil {
		res = p.fn(res)
	}
	go func() {
		p.mu.Lock()
		p.cache = res
		consumers := p.consumers
		p.mu.Unlock()
		for _, consumer := range consumers {
			consumer(res)
		}
	}()
}
