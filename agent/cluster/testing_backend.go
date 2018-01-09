package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"net/url"
)

type TestingBackendConfig struct {
	Consumer    bus.Consumer                           // Track consumer
	ReadyChan   chan struct{}                          // ready channel
	CrashChan   chan struct{}                          // Crash channel
	MessageChan chan map[string]map[string]interface{} // Messages
}

func NewTestingBackendFactory(backendConfig TestingBackendConfig) (f BackendFactory) {
	f = func(ctx context.Context, log *logx.Log, config Config) (b Backend, err error) {
		kvConfig := BackendConfig{
			Kind:    "local",
			Chroot:  "soil",
			ID:      config.NodeID,
			Address: "localhost",
			TTL:     config.TTL,
		}
		u, err := url.Parse(config.BackendURL)
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
			b = NewTestingBackend(ctx, log, backendConfig)
		}
		return
	}
	return
}

// Backend for testing purposes
type TestingBackend struct {
	*baseBackend
	config TestingBackendConfig
}

func NewTestingBackend(ctx context.Context, log *logx.Log, config TestingBackendConfig) (b *TestingBackend) {
	b = &TestingBackend{
		baseBackend: newBaseBackend(ctx, log, BackendConfig{}),
		config:      config,
	}
	go func() {
		select {
		case <-b.Ctx().Done():
		case <-config.ReadyChan:
			b.readyCancel()
		}
	}()
	go func() {
		select {
		case <-b.ctx.Done():
		case <-config.CrashChan:
			log.Trace(`crash`)
			b.config.Consumer.ConsumeMessage(bus.NewMessage("crash", map[string]interface{}{}))
			b.fail(fmt.Errorf(`crash`))
		}
	}()
	go func() {
		for {
			select {
			case <-b.ctx.Done():
				return
			case msg := <-b.config.MessageChan:
				for key, values := range msg {
					result := WatchResult{
						Key:  key,
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
			ID:      op.Message.Topic(),
			Hash:    op.Message.Payload().Hash(),
			WithTTL: op.WithTTL,
		})
		var res interface{}
		if err := op.Message.Payload().Unmarshal(&res); err != nil {
			b.log.Error(err)
			continue
		}
		data[op.Message.Topic()] = map[string]interface{}{
			"Data": res,
			"TTL":  op.WithTTL,
		}
	}
	select {
	case <-b.ctx.Done():
		b.log.Tracef(`skip send commits %v: backend closed`)
	case b.commitsChan <- commits:
		b.log.Tracef(`commits sent: %v`, commits)
		b.config.Consumer.ConsumeMessage(bus.NewMessage("test", data))
	}
}

func (b *TestingBackend) Subscribe(requests []WatchRequest) {
	for _, req := range requests {
		b.log.Tracef(`subscribe: %s`, req.Key)
	}
}
