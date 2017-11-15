package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/hashicorp/consul/api"
	"path"
	"time"
)

// Consul Backend
type ConsulBackend struct {
	*baseBackend

	conn      *api.Client
	sessionID string

	opsChan          chan []StoreOp
	watchRequestChan chan []WatchRequest
}

func NewConsulBackend(ctx context.Context, log *logx.Log, config BackendConfig) (w *ConsulBackend) {
	w = &ConsulBackend{
		baseBackend:      newBaseBackend(ctx, log, config),
		opsChan:          make(chan []StoreOp),
		watchRequestChan: make(chan []WatchRequest),
	}
	w.ctx, w.cancel = context.WithCancel(ctx)
	go w.connect()
	go w.loop()
	return
}

func (b *ConsulBackend) Close() error {
	b.cancel()
	return nil
}

func (b *ConsulBackend) Submit(op []StoreOp) {
	select {
	case <-b.ctx.Done():
		b.log.Warningf(`ignore %v: %v`, op, b.ctx.Err())
	case b.opsChan <- op:
		b.log.Tracef(`submit: %v`, op)
	}
}

func (b *ConsulBackend) Subscribe(req []WatchRequest) {
	select {
	case <-b.ctx.Done():
		b.log.Warningf(`ignore %v: %v`, req, b.ctx.Err())
	case b.watchRequestChan <- req:
		b.log.Tracef(`subscribe: %v`, req)
	}
}

func (b *ConsulBackend) loop() {
	b.log.Debug(`waiting for connect`)
	select {
	case <-b.ctx.Done():
		b.log.Error(`parent context is done on connect`)
		return
	case <-b.readyCtx.Done():
		b.log.Debug(`clean state reached`)
	}

LOOP:
	for {
		select {
		case <-b.ctx.Done():
			break LOOP
		case ops := <-b.opsChan:
			b.processStoreOpts(ops)
		case requests := <-b.watchRequestChan:
			for _, req := range requests {
				go b.watch(req)
			}
		}
	}
}

func (b *ConsulBackend) watch(req WatchRequest) {
	directory := path.Join(b.config.Chroot, req.Key)
	log := b.log.GetLog(b.log.Prefix(), append(b.log.Tags(), "watch", req.Key)...)
	log.Debugf(`open`)

	watchCtx, watchCancel := context.WithCancel(b.ctx)
	go func() {
		select {
		case <-req.Ctx.Done():
			watchCancel()
		case <-watchCtx.Done():
		}
	}()

	opts := (&api.QueryOptions{WaitTime: 15 * time.Second}).WithContext(watchCtx)
LOOP:
	for {
		select {
		case <-watchCtx.Done():
			break LOOP
		default:
		}
		pairs, meta, err := b.conn.KV().List(directory, opts)
		if err != nil {
			if err != context.Canceled {
				b.fail(err)
			}
			break LOOP
		}
		if opts.WaitIndex == meta.LastIndex {
			continue LOOP
		}
		opts.WaitIndex = meta.LastIndex

		result := WatchResult{
			Key: req.Key,
			Data: map[string][]byte{},
		}
		for _, pair := range pairs {
			result.Data[TrimKeyPrefix(directory, pair.Key)] = pair.Value
		}
		select {
		case <-watchCtx.Done():
			break LOOP
		case b.watchResultsChan <- result:
		}
	}
	log.Info(`closed`)
}

func (b *ConsulBackend) processStoreOpts(ops []StoreOp) {
	var kvOps api.KVTxnOps
	var commits []StoreCommit
	for _, op := range ops {
		key := path.Join(b.config.Chroot, op.Message.GetID())
		commits = append(commits, StoreCommit{
			ID:      op.Message.GetID(),
			Hash:    op.Message.Payload().Hash(),
			WithTTL: op.WithTTL,
		})
		if op.Message.Payload().IsEmpty() {
			kvOps = append(kvOps, &api.KVTxnOp{
				Verb: api.KVDelete,
				Key:  key,
			})
			continue
		}
		var value interface{}
		if err := op.Message.Payload().Unmarshal(&value); err != nil {
			b.log.Errorf(`can't unmarshal payload %s: %v`, op.Message.Payload(), err)
			continue
		}
		valJson, err := json.Marshal(value)
		if err != nil {
			b.log.Errorf(`can't marshal payload %v: %v`, value, err)
			continue
		}
		if op.WithTTL {
			kvOps = append(kvOps, &api.KVTxnOp{
				Verb:    api.KVLock,
				Key:     key,
				Session: b.sessionID,
				Value:   valJson,
			})
			continue
		}
		kvOps = append(kvOps, &api.KVTxnOp{
			Verb:  api.KVSet,
			Key:   key,
			Value: valJson,
		})
	}
	ok, _, _, txnErr := b.conn.KV().Txn(kvOps, (&api.QueryOptions{}).WithContext(b.ctx))
	if txnErr != nil {
		b.fail(txnErr)
		return
	}
	if !ok {
		b.fail(fmt.Errorf(`transaction failed`))
		return
	}
	select {
	case <-b.ctx.Done():
		b.log.Warningf(`skip to send commit for %v: %v`, commits, b.ctx.Err())
	case b.commitsChan <- commits:
		b.log.Debugf(`commits sent: %v`, commits)
	}
}

func (b *ConsulBackend) connect() {
	b.log.Debugf(`connecting: %s`, b.config.Address)
	var err error
	if b.conn, err = api.NewClient(&api.Config{
		Address: b.config.Address,
	}); err != nil {
		b.fail(err)
		return
	}
	// try to find session
	sessions, _, err := b.conn.Session().List((&api.QueryOptions{}).WithContext(b.ctx))
	if err != nil {
		b.fail(err)
		return
	}

	sessionName := path.Join(b.config.Chroot, b.config.ID)
	for _, session := range sessions {
		if session.Name == sessionName {
			b.sessionID = session.ID
			break
		}
	}
	if b.sessionID == "" {
		b.sessionID, _, err = b.conn.Session().Create(&api.SessionEntry{
			Name: b.config.ID,
			TTL:  b.config.TTL.String(),
		}, (&api.WriteOptions{}).WithContext(b.ctx))
		if err != nil {
			b.fail(err)
			return
		}
	}

	b.log.Infof(`connected`)
	b.readyCancel()

	// start watchdog
	go func() {
		b.log.Trace(`starting renew`)
		if renewErr := b.conn.Session().RenewPeriodic(b.config.TTL.String(), b.sessionID, (&api.WriteOptions{}).WithContext(b.ctx), nil); renewErr != nil && renewErr != context.Canceled {
			b.fail(fmt.Errorf(`renew failed: %v`, renewErr))
		}
		b.log.Trace(`renew closed`)
	}()
}
