package bus

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/supervisor"
	"sync"
)

type SimplePipe struct {
	consumers []MessageConsumer
	fn        func(message Message) Message
}

func NewSimplePipe(fn func(message Message) Message, consumers ...MessageConsumer) (p *SimplePipe) {
	p = &SimplePipe{
		fn:        fn,
		consumers: consumers,
	}
	return
}

func (p *SimplePipe) ConsumeMessage(message Message) {
	res := message
	if p.fn != nil {
		res = p.fn(res)
	}
	for _, consumer := range p.consumers {
		consumer.ConsumeMessage(res)
	}
}

// BoundedPipe registers on Dynamic producer on Open
type BoundedPipe struct {
	*supervisor.Control
	log      *logx.Log
	prefix   string
	producer BindableProducer
	fn       func(message Message) Message

	mu        *sync.Mutex
	cache     Message
	consumers []MessageConsumer
}

func NewPipe(ctx context.Context, log *logx.Log, prefix string, producer BindableProducer, fn func(message Message) Message, consumers ...MessageConsumer) (p *BoundedPipe) {
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
	go p.producer.RegisterConsumer(p.prefix, p)
	err = p.Control.Open()
	return
}

func (p *BoundedPipe) ConsumeMessage(message Message) {
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
			consumer.ConsumeMessage(res)
		}
	}()
}
