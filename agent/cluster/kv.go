package cluster

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/supervisor"
)

type kvConfigRequest struct {
	config   Config
	internal bool
}

type operatorConsumer struct {
	kv       *KV
	log      *logx.Log
	prefix   string
	volatile bool
}

func (c *operatorConsumer) ConsumeMessage(message bus.Message) (err error) {
	c.kv.Submit([]StoreOp{
		{
			Message: bus.NewMessage(NormalizeKey(c.prefix, message.GetID()), message.Payload()),
			WithTTL: c.volatile,
		},
	})
	return
}

type operatorProducer struct {
	key string
	kv  *KV
}

func (p *operatorProducer) Subscribe(ctx context.Context, consumer bus.Consumer) {
	p.kv.SubscribeKey(p.key, ctx, consumer)
}

type KV struct {
	*supervisor.Control
	log     *logx.Log
	factory BackendFactory

	backend Backend
	config  Config

	configRequestChan chan kvConfigRequest
	storeRequestsChan chan []StoreOp
	watchRequestsChan chan watcher

	volatile          map[string]bus.Message // volatile records
	pending           map[string]StoreOp     // pending ops
	commitsChan       chan []StoreCommit
	invokePendingChan chan struct{} // invoke pending operations

	registerWatchChan chan WatchRequest
	watchResultsChan  chan WatchResult

	watchGroups          map[string]*watchGroup
	pendingWatchGroups   map[string]struct{}
	closedWatchGroupChan chan string
}

func NewKV(ctx context.Context, log *logx.Log, factory BackendFactory) (b *KV) {
	b = &KV{
		Control:  supervisor.NewControl(ctx),
		log:      log.GetLog("cluster", "kv"),
		factory:  factory,
		config:   Config{},
		volatile: map[string]bus.Message{},
		pending:  map[string]StoreOp{},

		configRequestChan: make(chan kvConfigRequest, 1),
		storeRequestsChan: make(chan []StoreOp, 1),
		watchRequestsChan: make(chan watcher, 1),
		commitsChan:       make(chan []StoreCommit, 1),
		watchResultsChan:  make(chan WatchResult, 1),
		invokePendingChan: make(chan struct{}, 1),

		watchGroups:          map[string]*watchGroup{},
		pendingWatchGroups:   map[string]struct{}{},
		closedWatchGroupChan: make(chan string),
	}
	return
}

func (k *KV) Open() (err error) {
	go k.loop()
	err = k.Control.Open()
	return
}

func (k *KV) VolatileStore(prefix string) (consumer bus.Consumer) {
	consumer = &operatorConsumer{
		kv:       k,
		log:      k.log.GetLog("cluster", "kv", "store", "volatile", prefix),
		prefix:   prefix,
		volatile: true,
	}
	return
}

func (k *KV) PermanentStore(prefix string) (consumer bus.Consumer) {
	consumer = &operatorConsumer{
		kv:       k,
		log:      k.log.GetLog("cluster", "kv", "store", "permanent", prefix),
		prefix:   prefix,
		volatile: false,
	}
	return
}

func (k *KV) Producer(key string) (producer bus.Producer) {
	producer = &operatorProducer{
		kv:  k,
		key: key,
	}
	return
}

func (k *KV) Configure(config Config) {
	select {
	case <-k.Control.Ctx().Done():
		k.log.Warningf(`ignore config: %v`, k.Control.Ctx().Err())
	case k.configRequestChan <- kvConfigRequest{
		config: config,
	}:
		k.log.Tracef(`configure: %v`, config)
	}
}

// Submit store operations
func (k *KV) Submit(ops []StoreOp) {
	select {
	case <-k.Control.Ctx().Done():
		k.log.Warningf(`ignore submit: %v`, k.Control.Ctx().Err())
	case k.storeRequestsChan <- ops:
		k.log.Tracef(`submitted: %v`, ops)
	}
}

// Subscribe for changes
func (k *KV) SubscribeKey(key string, ctx context.Context, consumer bus.Consumer) {
	select {
	case <-k.Control.Ctx().Done():
		k.log.Warningf(`ignore subscribe: %v`, k.Control.Ctx().Err())
	case k.watchRequestsChan <- watcher{
		WatchRequest: WatchRequest{
			Key: key,
			Ctx: ctx,
		},
		consumer: consumer,
	}:
		k.log.Tracef(`subscribe: %s`, key)
	}
}

func (k *KV) loop() {
	log := k.log.GetLog("cluster", "kv", "loop")
	k.log.Info(`open`)
	k.backend = NewZeroBackend(k.Control.Ctx(), k.log)
	config := Config{}
LOOP:
	for {
		select {
		case <-k.Control.Ctx().Done():
			break LOOP
		case req := <-k.configRequestChan:
			log.Tracef(`received config: %v`, req)
			var needReconfigure bool
			select {
			case <-k.backend.Ctx().Done():
				needReconfigure = true
			default:
			}
			if !req.internal && !config.IsEqual(req.config) {
				log.Debugf(`external: %v->%v`, config, req.config)
				needReconfigure = true
			}
			if !needReconfigure {
				log.Tracef(`ignore reconfiguration`)
				continue LOOP
			}
			if config.NodeID != req.config.NodeID {
				log.Infof(`leaving cluster: node id changed %s->%s`, config.NodeID, req.config.NodeID)
				k.backend.Leave()
			} else {
				k.backend.Close()
			}
			var err error
			var backend Backend
			if backend, err = k.factory(k.Control.Ctx(), k.log, req.config); err != nil {
				k.log.Errorf(`can't create backend %v: %v'`, req.config, err)
				continue LOOP
			}
			newWatchdog(k, backend, req.config)
			config = req.config
			k.backend = backend
			k.log.Infof(`backend created: %v`, req.config)

			for id, message := range k.volatile {
				k.pending[id] = StoreOp{
					Message: message,
					WithTTL: true,
				}
			}
			for key := range k.watchGroups {
				k.pendingWatchGroups[key] = struct{}{}
			}
		case ops := <-k.storeRequestsChan:
			log.Tracef(`submit: %v`, ops)
			for _, op := range ops {
				id := op.Message.GetID()
				if op.WithTTL {
					// volatile
					if op.Message.Payload().IsEmpty() {
						delete(k.volatile, id)
					} else {
						k.volatile[id] = op.Message
					}
				}
				k.pending[id] = op
			}
			go func() {
				select {
				case <-k.Control.Ctx().Done():
				case k.invokePendingChan <- struct{}{}:
				}
			}()
		case <-k.invokePendingChan:
			log.Tracef(`invoke: (pending: %v, watch: %v)`, k.pending, k.pendingWatchGroups)
			select {
			case <-k.backend.ReadyCtx().Done():
				select {
				case <-k.backend.Ctx().Done():
					log.Trace(`skip send pending: backend is closed`)
				default:
					if len(k.pending) > 0 {
						var ops []StoreOp
						for _, op := range k.pending {
							ops = append(ops, op)
						}
						k.backend.Submit(ops)
						log.Tracef(`submitted: %v`, ops)
					}
					if len(k.pendingWatchGroups) > 0 {
						var requests []WatchRequest
						for key := range k.pendingWatchGroups {
							if group, ok := k.watchGroups[key]; ok {
								requests = append(requests, WatchRequest{
									Key: group.key,
									Ctx: group.ctx,
								})
							}
							delete(k.pendingWatchGroups, key)
						}
						k.backend.Subscribe(requests)
					}
				}
			default:
				log.Trace(`skip send pending: backend is not ready`)
			}
		case commits := <-k.commitsChan:
			log.Tracef(`commits: %v`, commits)
			select {
			case <-k.backend.ReadyCtx().Done():
				// ok send ops
				select {
				case <-k.backend.Ctx().Done():
					log.Trace(`skip commit: backend is closed`)
				default:
					for _, commit := range commits {
						delete(k.pending, commit.ID)
					}
					log.Tracef(`commits done: %v (pending %v)`, commits, k.pending)
				}
			default:
				log.Trace(`skip commit: backend is not ready`)
			}
		case req := <-k.watchRequestsChan:
			if group, ok := k.watchGroups[req.Key]; ok {
				k.log.Tracef(`watch group %s found: adding watcher`, group.key)
				group.register(req)
				continue LOOP
			}
			k.log.Tracef(`creating watching group for %s`, req.Key)
			group := newWatchGroup(k.Control.Ctx(), k.log, req.Key)
			k.watchGroups[req.Key] = group
			group.register(req)
			go func() {
				select {
				case <-k.Control.Ctx().Done():
				case <-group.ctx.Done():
					k.closedWatchGroupChan <- req.Key
				}
			}()
			k.pendingWatchGroups[req.Key] = struct{}{}
			go func() {
				select {
				case <-k.Control.Ctx().Done():
				case k.invokePendingChan <- struct{}{}:
				}
			}()
		case result := <-k.watchResultsChan:
			log.Tracef(`watch result: %v`, result)
			if group, ok := k.watchGroups[result.Key]; ok {
				group.AcceptResult(result)
				k.log.Tracef(`message %v sent to watch group`, result)
				continue LOOP
			}
			k.log.Warningf(`watch group %s is not found`, result.Key)
		case id := <-k.closedWatchGroupChan:
			delete(k.watchGroups, id)
		}
	}
	k.log.Info(`close`)
}
