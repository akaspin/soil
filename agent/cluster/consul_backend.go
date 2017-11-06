package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/hashicorp/consul/api"
	"path"
	"strings"
	"time"
)

// Consul Backend
type ConsulBackend struct {

	conn      *api.Client
	sessionID string

	*baseBackend


	opsChan          chan []BackendStoreOp
	watchRequestChan chan []BackendWatchRequest
}

func NewConsulBackend(ctx context.Context, log *logx.Log, config BackendConfig) (w *ConsulBackend) {
	w = &ConsulBackend{
		baseBackend: newBaseBackend(ctx, log, config),
		opsChan:          make(chan []BackendStoreOp, 1),
		watchRequestChan: make(chan []BackendWatchRequest, 1),
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

func (b *ConsulBackend) Submit(op []BackendStoreOp) {
	select {
	case <-b.ctx.Done():
		b.log.Warningf(`ignore %v: %v`, op, b.ctx.Err())
	case b.opsChan <- op:
		b.log.Tracef(`submit: %v`, op)
	}
}

// Subscribe
func (b *ConsulBackend) Subscribe(req []BackendWatchRequest) {
	select {
	case <-b.ctx.Done():
		b.log.Warningf(`ignore %v: %v`, req, b.ctx.Err())
	case b.watchRequestChan <- req:
		b.log.Tracef(`subscribe: %v`, req)
	}
}

func (b *ConsulBackend) WatchChan() chan bus.Message {
	return b.watchChan
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
			b.log.Tracef(`context done`)
			break LOOP
		case ops := <-b.opsChan:
			b.log.Tracef(`received: %v`, ops)
			b.handleStoreOps(ops)
		case requests := <-b.watchRequestChan:
			b.log.Tracef(`received watch: %v`, requests)
			for _, req := range requests {
				go b.watch(req)
			}
		}
	}
}

func (b *ConsulBackend) watch(req BackendWatchRequest) {
	directory := path.Join(b.config.Chroot, req.Key)
	log := b.log.GetLog(b.log.Prefix(), append(b.log.Tags(), "watch", req.Key)...)
	log.Debugf(`open`)

	watchCtx, watchCancel := context.WithCancel(b.ctx)
	// bound watch context to request
	go func() {
		select {
		case <-req.Ctx.Done():
			watchCancel()
		case <-watchCtx.Done():
			// also terminate on watch ctx done
		}
	}()

	opts := (&api.QueryOptions{WaitTime: 15 * time.Second}).WithContext(watchCtx)
LOOP:
	for {
		select {
		case <-watchCtx.Done():
			log.Trace(b.ctx.Err())
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

		// pack pairs
		values := map[string]interface{}{}
		for _, pair := range pairs {
			var value interface{}
			if jsonErr := json.NewDecoder(bytes.NewReader(pair.Value)).Decode(&value); jsonErr != nil {
				log.Error(err)
				continue
			}
			values[strings.TrimPrefix(pair.Key, directory+"/")] = value
		}
		select {
		case <-watchCtx.Done():
			log.Trace(b.ctx.Err())
			break LOOP
		case b.watchChan <- bus.NewMessage(req.Key, values):
			log.Tracef(`sent: %v`, values)
		}
	}

}

func (b *ConsulBackend) handleStoreOps(ops []BackendStoreOp) {
	var kvOps api.KVTxnOps
	var commits []BackendCommit
	for _, op := range ops {
		key := path.Join(b.config.Chroot, op.Message.GetID())
		commits = append(commits, BackendCommit{
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
		valJson, err := op.Message.Payload().JSON()
		if err != nil {
			b.fail(err)
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

func (b *ConsulBackend) fail(err error) {
	b.log.Errorf(`failed: %v`, err)
	b.cancel()
}
