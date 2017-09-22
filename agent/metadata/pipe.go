package metadata

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/supervisor"
	"sync"
)

type SimplePipe struct {
	consumers []func(message Message)
	fn        func(message Message) Message
}

func NewSimplePipe(fn func(message Message) Message, consumers ...func(message Message)) (p *SimplePipe) {
	p = &SimplePipe{
		fn:        fn,
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

// BoundedPipe registers on Dynamic producer on Open
type BoundedPipe struct {
	*supervisor.Control
	log      *logx.Log
	prefix   string
	producer DynamicProducer
	fn       func(message Message) Message

	mu        *sync.Mutex
	cache     Message
	consumers []func(message Message)
}

func NewPipe(ctx context.Context, log *logx.Log, prefix string, producer DynamicProducer, fn func(message Message) Message, consumers ...func(message Message)) (p *BoundedPipe) {
	p = &BoundedPipe{
		Control:   supervisor.NewControl(ctx),
		log:       log.GetLog("pipe", prefix),
		prefix:    prefix,
		producer:  producer,
		fn:        fn,
		mu:        &sync.Mutex{},
		consumers: consumers,
	}
	return
}

func (p *BoundedPipe) Open() (err error) {
	go p.producer.RegisterConsumer(p.prefix, p.sync)
	err = p.Control.Open()
	return
}

func (p *BoundedPipe) sync(message Message) {
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
