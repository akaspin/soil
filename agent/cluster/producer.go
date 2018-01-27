package cluster

import (
	"context"
	"github.com/akaspin/soil/agent/bus"
)

type producer struct {
	key string
	kv  *KV
}

func NewProducer(kv *KV, key string) (p bus.Producer) {
	return &producer{
		key: key,
		kv:  kv,
	}
}

func (p *producer) Subscribe(ctx context.Context, consumer bus.Consumer) {
	p.kv.SubscribeKey(p.key, ctx, consumer)
}
