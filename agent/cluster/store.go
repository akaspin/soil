package cluster

import (
	"github.com/akaspin/soil/agent/bus"
	"path"
)

type Store struct {
	kv       *KV
	volatile bool
	prefix   string
}

func NewVolatileStore(kv *KV, prefix string) (s bus.Consumer) {
	s = &Store{
		kv:       kv,
		prefix:   prefix,
		volatile: true,
	}
	return
}

func NewPermanentStore(kv *KV, prefix string) (s bus.Consumer) {
	s = &Store{
		kv:       kv,
		prefix:   prefix,
		volatile: false,
	}
	return
}

func (o *Store) ConsumeMessage(message bus.Message) {
	o.kv.Submit([]StoreOp{
		{
			Message: bus.NewMessage(path.Join(o.prefix, message.GetID()), message.Payload()),
			WithTTL: o.volatile,
		},
	})
}
