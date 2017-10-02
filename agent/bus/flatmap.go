package bus

import (
	"sync"
)

type FlatMap struct {
	//*supervisor.Control
	//log *logx.Log

	prefix    string
	isStrict  bool
	consumers []MessageConsumer

	mu     sync.Mutex
	cached Message
}

func NewFlatMap(strict bool, prefix string, consumers ...MessageConsumer) (p *FlatMap) {
	p = &FlatMap{
		//Control:   supervisor.NewControl(ctx),
		//log:       log.GetLog("producer", prefix),
		prefix:    prefix,
		isStrict:  strict,
		consumers: consumers,
		cached:    NewMessage(prefix, nil),
	}
	return
}

//func (p *FlatMap) Open() (err error) {
//	p.log.Debug("open")
//	err = p.Control.Open()
//	return
//}
//
//func (p *FlatMap) Close() error {
//	p.log.Debug("close")
//	return p.Control.Close()
//}

// Set specific keys
func (p *FlatMap) Set(data map[string]string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.isStrict {
		p.cached = NewMessage(p.prefix, data)
	} else {
		chunk := CloneMap(p.cached.GetPayload())
		for k, v := range data {
			chunk[k] = v
		}
		p.cached = NewMessage(p.prefix, chunk)
	}
	p.notifyAll()
}

func (p *FlatMap) notifyAll() {
	//p.log.Tracef("syncing with %d consumers", len(p.consumers))
	for _, consumer := range p.consumers {
		consumer.ConsumeMessage(p.cached)
	}
}
