package cluster

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/supervisor"
	"time"
)

type KV struct {
	*supervisor.Control
	log     *logx.Log
	factory BackendFactory

	backend Backend
	config  Config

	volatile map[string]bus.Message    // volatile records
	pending  map[string]BackendStoreOp // pending ops

	configRequestChan chan configRequest
	submitChan        chan []BackendStoreOp
	commitsChan       chan []BackendCommit

	invokeChan chan struct{}
}

func NewKV(ctx context.Context, log *logx.Log, factory BackendFactory) (b *KV) {
	b = &KV{
		Control:  supervisor.NewControl(ctx),
		log:      log.GetLog("cluster", "backend"),
		factory:  factory,
		config:   Config{},
		volatile: map[string]bus.Message{},
		pending:  map[string]BackendStoreOp{},

		configRequestChan: make(chan configRequest, 1),
		submitChan:        make(chan []BackendStoreOp, 1),
		commitsChan:       make(chan []BackendCommit, 1),

		invokeChan: make(chan struct{}, 1),
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
	case k.configRequestChan <- configRequest{
		config: config,
	}:
		k.log.Tracef(`configure: %v`, config)
	}
}

func (k *KV) Submit(ops []BackendStoreOp) {
	select {
	case <-k.Control.Ctx().Done():
		k.log.Warningf(`ignore submit: %v`, k.Control.Ctx().Err())
	case k.submitChan <- ops:
		k.log.Tracef(`submitted: %v`, ops)
	}
}

func (k *KV) Subscribe(key string, ctx context.Context, consumer bus.Consumer) {

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
				k.pending[id] = BackendStoreOp{
					Message: message,
					WithTTL: true,
				}
			}
			k.backend = k.createBackend(req.config)
			log.Debugf(`created: %v`, k.config)
		case ops := <-k.submitChan:
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
				case k.invokeChan <- struct{}{}:
				}
			}()
		case <-k.invokeChan:
			log.Tracef(`invoke: (pending: %v)`, k.pending)
			select {
			case <-k.backend.ReadyCtx().Done():
				// ok send ops
				select {
				case <-k.backend.Ctx().Done():
					log.Trace(`skip send pending: backend is closed`)
				default:
					if len(k.pending) == 0 {
						log.Trace(`skip submit: pending is empty`)
						continue LOOP
					}
					var ops []BackendStoreOp
					for _, op := range k.pending {
						ops = append(ops, op)
					}
					k.backend.Submit(ops)
					log.Tracef(`submitted: %v`, ops)
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
		}
	}
}

func (k *KV) createBackend(config Config) (backend Backend) {
	var kvErr error
	if backend, kvErr = k.factory(k.Control.Ctx(), k.log, config); kvErr != nil {
		k.log.Error(kvErr)
	}
	newBackendWatchdog(k, backend, config)

	k.log.Debugf(`created: %v`, config)
	return
}

// Backend watchdog evaluates Backend contexts and commit channel
type backendWatchdog struct {
	kv      *KV
	backend Backend
	config  Config
	log     *logx.Log
}

func newBackendWatchdog(kv *KV, backend Backend, config Config) (w *backendWatchdog) {
	w = &backendWatchdog{
		kv:      kv,
		backend: backend,
		config:  config,
		log:     kv.log.GetLog("cluster", "watchdog", config.URL, config.ID),
	}
	go w.ready()
	go w.done()
	go w.commit()
	return
}

// watch ready context
func (w *backendWatchdog) ready() {
	select {
	case <-w.backend.Ctx().Done():
		return
	case <-w.backend.ReadyCtx().Done():
		w.log.Trace(`backend is ready`)
		select {
		case <-w.kv.Control.Ctx().Done():
		case w.kv.invokeChan <- struct{}{}:
			w.log.Debug(`try request sent`)
		}
	}
}

func (w *backendWatchdog) done() {
	<-w.backend.Ctx().Done()
	w.log.Tracef(`backend closed: sending wake request after %s`, w.config.RetryInterval)
	select {
	case <-w.kv.Control.Ctx().Done():
		return
	case <-time.After(w.config.RetryInterval):
		w.log.Trace(`sending wake request`)
		select {
		case <-w.kv.Control.Ctx().Done():
			w.log.Tracef(`skip: master context closed`)
			return
		case w.kv.configRequestChan <- configRequest{
			config:   w.config,
			internal: true,
		}:
			w.log.Trace(`wake request sent`)
		}
	}
}

func (w *backendWatchdog) commit() {
	w.log.Trace(`commits: open`)
LOOP:
	for {
		select {
		case <-w.backend.Ctx().Done():
			break LOOP
		case commits := <-w.backend.CommitChan():
			w.log.Tracef(`received commits: %v`, commits)
			select {
			case <-w.backend.Ctx().Done():
				w.log.Tracef(`skip sending commits %v: backend closed`)
				break LOOP
			case w.kv.commitsChan <- commits:
				w.log.Tracef(`commits sent: %v`, commits)
			}
		}
	}
	w.log.Trace(`commit: close`)
}

type configRequest struct {
	config   Config
	internal bool
}
