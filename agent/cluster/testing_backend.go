package cluster

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"net/url"
	"encoding/json"
)

func NewTestingBackendFactory(consumer bus.Consumer, crashChan chan struct{}, msgChan chan map[string]map[string]interface{}) (f BackendFactory) {
	f = func(ctx context.Context, log *logx.Log, config Config) (b Backend, err error) {
		kvConfig := BackendConfig{
			Kind:    "local",
			Chroot:  "soil",
			ID:      config.ID,
			Address: "localhost",
			TTL:     config.TTL,
		}
		u, err := url.Parse(config.URL)
		if err != nil {
			log.Error(err)
		}
		if err == nil {
			kvConfig.Kind = u.Scheme
			kvConfig.Address = u.Host
			kvConfig.Chroot = u.Path
		}
		kvLog := log.GetLog("cluster", "backend", kvConfig.Kind)
		switch kvConfig.Kind {
		case "zero":
			b = NewZeroBackend(ctx, kvLog)
		default:
			b = NewTestingBackend(ctx, log, consumer, crashChan, msgChan)
		}
		return
	}
	return
}

// Backend for testing purposes
type TestingBackend struct {
	*baseBackend
	consumer bus.Consumer
	msgChan  chan map[string]map[string]interface{}
}

func NewTestingBackend(ctx context.Context, log *logx.Log, consumer bus.Consumer, crashChan chan struct{}, msgChan chan map[string]map[string]interface{}) (b *TestingBackend) {
	b = &TestingBackend{
		baseBackend: newBaseBackend(ctx, log, BackendConfig{}),
		consumer:    consumer,
		msgChan:     msgChan,
	}
	b.readyCancel()
	go func() {
		select {
		case <-b.ctx.Done():
		case <-crashChan:
			log.Trace(`crash`)
			b.fail(fmt.Errorf(`crash`))
		}
	}()
	go func() {
		for {
			select {
			case <-b.ctx.Done():
				return
			case msg := <-b.msgChan:
				for key, values := range msg {
					result := WatchResult{
						Key: key,
						Data: map[string][]byte{},
					}
					for k, v := range values {
						buf, err := json.Marshal(v)
						if err != nil {
							log.Error(err)
							continue
						}
						result.Data[k] = buf
					}
					select {
					case <-b.ctx.Done():
						return
					case b.watchResultsChan <- result:
					}
				}
			}
		}
	}()
	return
}

func (b *TestingBackend) Submit(ops []StoreOp) {
	data := map[string]interface{}{}
	var commits []StoreCommit
	for _, op := range ops {
		commits = append(commits, StoreCommit{
			ID:      op.Message.GetID(),
			Hash:    op.Message.Payload().Hash(),
			WithTTL: op.WithTTL,
		})
		var res interface{}
		if err := op.Message.Payload().Unmarshal(&res); err != nil {
			b.log.Error(err)
			continue
		}
		data[op.Message.GetID()] = map[string]interface{}{
			"Data": res,
			"TTL":  op.WithTTL,
		}
	}
	select {
	case <-b.ctx.Done():
		b.log.Tracef(`skip send commits %v: backend closed`)
	case b.commitsChan <- commits:
		b.log.Tracef(`commits sent: %v`, commits)
		b.consumer.ConsumeMessage(bus.NewMessage("test", data))
	}
}

func (b *TestingBackend) Subscribe(requests []WatchRequest) {
	for _, req := range requests {
		b.log.Tracef(`subscribe: %s`, req.Key)
	}
}
