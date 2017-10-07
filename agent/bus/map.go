package bus

import (
	"sync"
	"github.com/mitchellh/hashstructure"
	"sync/atomic"
)

type MapUpstream struct {
	name string
	consumers []MessageConsumer
	mu     sync.Mutex
	cache map[string]string
	mark uint64
}


func NewMapUpstream(name string, consumers ...MessageConsumer) (u *MapUpstream) {
	u = &MapUpstream{
		name: name,
		consumers: consumers,
		cache: map[string]string{},
		mark: ^uint64(0),
	}
	return
}

func (u *MapUpstream) Set(data map[string]string) {
	u.mu.Lock()
	for k, v := range data {
		u.cache[k] = v
	}
	var needNotify bool
	var message Message
	newMark, _ := hashstructure.Hash(u.cache, nil)
	if newMark != u.mark {
		u.mark = newMark
		needNotify = true
		message = NewMessage(u.name, u.cache)
	}
	u.mu.Unlock()
	if needNotify {
		u.notifyAll(message)
	}
}

func (u *MapUpstream) Delete(key ...string) {
	u.mu.Lock()
	for _, k := range key {
		delete(u.cache, k)
	}
	var needNotify bool
	var message Message
	newMark, _ := hashstructure.Hash(u.cache, nil)
	if newMark != u.mark {
		u.mark = newMark
		needNotify = true
		message = NewMessage(u.name, u.cache)
	}
	u.mu.Unlock()
	if needNotify {
		u.notifyAll(message)
	}
}

func (u *MapUpstream) notifyAll(message Message) {
	for _, consumer := range u.consumers {
		consumer.ConsumeMessage(message)
	}
}

type StrictMapUpstream struct {
	consumers []MessageConsumer
	name string
	mu     sync.Mutex
	mark uint64
}

func NewStrictMapUpstream(name string, consumers ...MessageConsumer) (p *StrictMapUpstream) {
	p = &StrictMapUpstream{
		name: name,
		consumers: consumers,
		mark: ^uint64(0),
	}
	return
}

// Set specific keys
func (u *StrictMapUpstream) Set(data map[string]string) {
	message := NewMessage(u.name, data)
	if atomic.SwapUint64(&u.mark, message.GetMark()) != message.GetMark() {
		for _, consumer := range u.consumers {
			consumer.ConsumeMessage(message)
		}
	}
}
