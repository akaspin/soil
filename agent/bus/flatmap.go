package bus

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/supervisor"
	"sync"
)

type FlatMap struct {
	*supervisor.Control
	log *logx.Log

	prefix    string
	isStrict  bool
	consumers []MessageConsumer

	mu   *sync.Mutex
	data map[string]string
}

func NewFlatMap(ctx context.Context, log *logx.Log, strict bool, prefix string, consumers ...MessageConsumer) (p *FlatMap) {
	p = &FlatMap{
		Control:   supervisor.NewControl(ctx),
		log:       log.GetLog("producer", prefix),
		prefix:    prefix,
		isStrict:  strict,
		consumers: consumers,
		mu:        &sync.Mutex{},
		data:      map[string]string{},
	}
	return
}

func (p *FlatMap) Open() (err error) {
	p.log.Debug("open")
	err = p.Control.Open()
	return
}

func (p *FlatMap) Close() error {
	p.log.Debug("close")
	return p.Control.Close()
}

// Set specific keys
func (p *FlatMap) Set(data map[string]string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.isStrict {
		p.data = data
	} else {
		for k, v := range data {
			p.data[k] = v
		}
	}
	p.notifyAll()
}

func (p *FlatMap) notifyAll() {
	p.log.Tracef("syncing with %d consumers", len(p.consumers))
	msg := NewMessage(p.prefix, p.data)
	for _, consumer := range p.consumers {
		consumer.ConsumeMessage(msg)
	}
}
