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

	registerWatchChan    chan WatchRequest
	watchResultsChan     chan bus.Message
	watchGroups          map[string]*watchGroup
	pendingWatchGroups   map[string]struct{}
	closedWatchGroupChan chan string
}

func NewKV(ctx context.Context, log *logx.Log, factory BackendFactory) (b *KV) {
	b = &KV{
		Control:  supervisor.NewControl(ctx),
		log:      log.GetLog("cluster", "backend"),
		factory:  factory,
		config:   Config{},
		volatile: map[string]bus.Message{},
		pending:  map[string]StoreOp{},

		configRequestChan: make(chan kvConfigRequest),
		storeRequestsChan: make(chan []StoreOp),
		watchRequestsChan: make(chan watcher),
		commitsChan:       make(chan []StoreCommit),
		watchResultsChan:  make(chan bus.Message),
		invokePendingChan: make(chan struct{}),

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
func (k *KV) Subscribe(key string, ctx context.Context, consumer bus.Consumer) {
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
	log := k.log.GetLog("cluster", "backend", "loop")
	k.backend = NewZeroBackend(k.Control.Ctx(), k.log)
LOOP:
	for {
		select {
		case <-k.Control.Ctx().Done():
			log.Debugf(`control closed`)
			break LOOP
		case req := <-k.configRequestChan:
			log.Tracef(`received config: %v`, req)
			var needReconfigure bool
			select {
			case <-k.backend.Ctx().Done():
				needReconfigure = true
			default:
			}
			if !req.internal && !k.config.IsEqual(req.config) {
				log.Debugf(`external: %v->%v`, k.config, req.config)
				k.backend.Close()
				needReconfigure = true
			}
			if !needReconfigure {
				log.Tracef(`ignore reconfiguration`)
				continue LOOP
			}
			k.config = req.config
			for id, message := range k.volatile {
				k.pending[id] = StoreOp{
					Message: message,
					WithTTL: true,
				}
			}
			for key := range k.watchGroups {
				k.pendingWatchGroups[key] = struct{}{}
			}
			k.backend = k.createBackend(req.config)
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
				// ok send ops
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
						var watchReqs []WatchRequest
						for key := range k.pendingWatchGroups {
							if group, ok := k.watchGroups[key]; ok {
								watchReqs = append(watchReqs, WatchRequest{
									Key: group.key,
									Ctx: group.ctx,
								})
							}
							delete(k.pendingWatchGroups, key)
						}
						k.backend.Subscribe(watchReqs)
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
		case res := <-k.watchResultsChan:
			log.Tracef(`watch result: %v`, res)
			if group, ok := k.watchGroups[res.GetID()]; ok {
				group.ConsumeMessage(res)
				k.log.Tracef(`message %v sent to watch group`, res)
				continue LOOP
			}
			k.log.Warningf(`watch group %s is not found`, res.GetID())
		case id := <-k.closedWatchGroupChan:
			delete(k.watchGroups, id)
		}
	}
}

func (k *KV) createBackend(config Config) (backend Backend) {
	var kvErr error
	if backend, kvErr = k.factory(k.Control.Ctx(), k.log, config); kvErr != nil {
		k.log.Error(kvErr)
	}
	newWatchdog(k, backend, config)

	k.log.Debugf(`created: %v`, config)
	return
}
