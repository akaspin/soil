package cluster

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/hashicorp/consul/api"
)

type ConsulWorker struct {
	ctx    context.Context
	cancel context.CancelFunc
	log    *logx.Log
	config WorkerConfig

	conn      *api.Client
	sessionID string

	cleanCtx      context.Context
	cleanCancel   context.CancelFunc
	failureCtx    context.Context
	failureCancel context.CancelFunc

	opsChan    chan []WorkerStoreOp
	commitChan chan []string
}

func NewConsulWorker(ctx context.Context, log *logx.Log, config WorkerConfig) (w *ConsulWorker) {
	w = &ConsulWorker{
		log:        log,
		config:     config,
		opsChan:    make(chan []WorkerStoreOp, 1),
		commitChan: make(chan []string, 1),
	}
	w.ctx, w.cancel = context.WithCancel(ctx)
	w.cleanCtx, w.cleanCancel = context.WithCancel(context.Background())
	w.failureCtx, w.failureCancel = context.WithCancel(context.Background())
	go w.connect()
	go w.loop()
	return
}

func (w *ConsulWorker) Close() error {
	w.cancel()
	return nil
}

func (w *ConsulWorker) Submit(op []WorkerStoreOp) {
	select {
	case <-w.ctx.Done():
		w.log.Warningf(`ignore %v: %v`, op, w.ctx.Err())
	case w.opsChan <- op:
		w.log.Tracef(`submit: %v`, op)
	}
}

func (w *ConsulWorker) CleanCtx() context.Context {
	return w.cleanCtx
}

func (w *ConsulWorker) FailureCtx() context.Context {
	return w.failureCtx
}

func (w *ConsulWorker) CommitChan() chan []string {
	return w.commitChan
}

func (w *ConsulWorker) loop() {
	w.log.Debug(`waiting for connect`)
	select {
	case <-w.ctx.Done():
		w.log.Error(`parent context is done on connect`)
		return
	case <-w.cleanCtx.Done():
		w.log.Debug(`clean state reached`)
	}

LOOP:
	for {
		select {
		case <-w.ctx.Done():
			w.log.Tracef(`context done`)
			break LOOP
		case ops := <-w.opsChan:
			w.log.Tracef(`received: %v`, ops)
			w.handleStoreOps(ops)
		}
	}
}

func (w *ConsulWorker) handleStoreOps(ops []WorkerStoreOp) {
	var kvOps api.KVTxnOps
	var keys []string
	for _, op := range ops {
		keys = append(keys, op.Message.GetID())
		if op.Message.Payload().IsEmpty() {
			kvOps = append(kvOps, &api.KVTxnOp{
				Verb: api.KVDelete,
				Key:  op.Message.GetID(),
			})
			continue
		}
		valJson, err := op.Message.Payload().JSON()
		if err != nil {
			w.fail(err)
		}
		if op.WithTTL {
			kvOps = append(kvOps, &api.KVTxnOp{
				Verb:    api.KVLock,
				Key:     op.Message.GetID(),
				Session: w.sessionID,
				Value:   valJson,
			})
			continue
		}
		kvOps = append(kvOps, &api.KVTxnOp{
			Verb:  api.KVSet,
			Key:   op.Message.GetID(),
			Value: valJson,
		})
	}
	ok, _, _, txnErr := w.conn.KV().Txn(kvOps, (&api.QueryOptions{}).WithContext(w.ctx))
	if txnErr != nil {
		w.fail(txnErr)
		return
	}
	if !ok {
		w.fail(fmt.Errorf(`transaction failed`))
		return
	}
	select {
	case <-w.ctx.Done():
		w.log.Warningf(`skip to send commit for %v: %v`, keys, w.ctx.Err())
	case w.commitChan <- keys:
		w.log.Debugf(`commits sent: %v`, keys)
	}
}

func (w *ConsulWorker) connect() {
	w.log.Debugf(`connecting: %s`, w.config.Address)
	var err error
	if w.conn, err = api.NewClient(&api.Config{
		Address: w.config.Address,
	}); err != nil {
		w.fail(err)
		return
	}
	// try to find session
	sessions, _, err := w.conn.Session().List((&api.QueryOptions{}).WithContext(w.ctx))
	if err != nil {
		w.fail(err)
		return
	}

	for _, session := range sessions {
		if session.Name == w.config.ID {
			w.sessionID = session.ID
			break
		}
	}
	if w.sessionID == "" {
		w.sessionID, _, err = w.conn.Session().Create(&api.SessionEntry{
			Name: w.config.ID,
			TTL:  w.config.TTL.String(),
		}, (&api.WriteOptions{}).WithContext(w.ctx))
		if err != nil {
			w.fail(err)
			return
		}
	}

	w.log.Infof(`connected`)
	w.cleanCancel()

	// start watchdog
	go func() {
		w.log.Trace(`starting renew`)
		if renewErr := w.conn.Session().RenewPeriodic(w.config.TTL.String(), w.sessionID, (&api.WriteOptions{}).WithContext(w.ctx), nil); renewErr != nil && renewErr != context.Canceled {
			w.fail(fmt.Errorf(`renew failed: %v`, renewErr))
		}
		w.log.Trace(`renew closed`)
	}()
}

func (w *ConsulWorker) fail(err error) {
	w.log.Errorf(`failed: %v`, err)
	w.cancel()
	w.failureCancel()
}
