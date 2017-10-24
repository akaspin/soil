package backend

import (
	"context"
	"errors"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/supervisor"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"github.com/docker/libkv/store/zookeeper"
	"github.com/mitchellh/hashstructure"
	"net/url"
	"path"
	"strings"
	"time"
)

var disabledError = errors.New("public namespace is disabled")

const (
	opSetTTL = iota
	opSet
	opDelete
)

type operation struct {
	key   string
	value string
	op    int
}

type Options struct {
	Enabled       bool
	Advertise     string
	URL           string
	Timeout       time.Duration
	RetryInterval time.Duration
	TTL           time.Duration
}

type LibKVBackend struct {
	*supervisor.Control
	log     *logx.Log
	options Options

	// connection context
	connDirtyCtx    context.Context
	connDirtyCancel context.CancelFunc
	conn            store.Store
	connErr         error
	chroot          string
	ttl             time.Duration

	operationChan chan []operation
}

func NewLibKVBackend(ctx context.Context, log *logx.Log, options Options) (b *LibKVBackend) {
	b = &LibKVBackend{
		Control:       supervisor.NewControl(ctx),
		log:           log.GetLog("public", "backend"),
		options:       options,
		operationChan: make(chan []operation, 500),
	}
	b.connDirtyCtx, b.connDirtyCancel = context.WithCancel(context.Background())
	return
}

func (b *LibKVBackend) Open() (err error) {
	go b.connect()
	go b.operationLoop()
	err = b.Control.Open()
	return
}

// RegisterConsumer Registers Consumer with specific prefix
func (b *LibKVBackend) RegisterConsumer(prefix string, consumer bus.Consumer) {
	go b.watchLoop(prefix, consumer)
}

// delete data
func (b *LibKVBackend) deleteWithPrefix(prefix string, keys ...string) {
	if !b.options.Enabled {
		return
	}
	var ops []operation
	for _, key := range keys {
		ops = append(ops, operation{
			key: path.Join(prefix, key),
			op:  opDelete,
		})
	}
	go func() {
		b.operationChan <- ops
	}()
}

// Set data in storage
func (b *LibKVBackend) set(prefix string, data map[string]string, withTTL bool) {
	if !b.options.Enabled {
		return
	}
	var ops []operation
	op := opSet
	if withTTL {
		op = opSetTTL
	}
	for key, value := range data {
		ops = append(ops, operation{
			key:   path.Join(prefix, key),
			value: value,
			op:    op,
		})
	}
	go func() {
		b.operationChan <- ops
	}()
}

func (b *LibKVBackend) connect() {
	defer b.connDirtyCancel()

	if !b.options.Enabled {
		b.log.Info("disabled")
		b.connErr = disabledError
		return
	}

	u, err := url.Parse(b.options.URL)
	if err != nil {
		b.log.Error(err)
		b.connErr = err
		return
	}
	kind := store.Backend(u.Scheme)
	addr := strings.Split(u.Host, ",")
	b.chroot = strings.TrimPrefix(u.Path, "/")
	b.ttl = b.options.TTL

	switch kind {
	case store.CONSUL:
		// libkv divides TTL by 2
		b.ttl = b.ttl * 2
		consul.Register()
	case store.ETCD:
		etcd.Register()
	case store.ZK:
		zookeeper.Register()
	default:
		err = fmt.Errorf("invalid backend type: %s", kind)
		return
	}

	var retry int
	for {
		retry++
		b.log.Infof("connecting to %s (retry %d)", b.options.URL, retry)
		b.conn, b.connErr = libkv.NewStore(kind, addr, &store.Config{
			ConnectionTimeout: b.options.Timeout,
		})
		if b.connErr != nil {
			b.log.Errorf("failed to connect to %s: %v: sleeping %v", b.options.URL, b.connErr, b.options.RetryInterval)
			time.Sleep(b.options.RetryInterval)
			continue
		} else {
			b.log.Infof("connected to %s", b.options.URL)
			return
		}
	}
}

func (b *LibKVBackend) operationLoop() {
	log := b.log.GetLog(b.log.Prefix(), "operation")
	defer log.Info("close")

	if !b.options.Enabled {
		log.Info("disabled")
		return
	}

	log.Infof("open: TTL %v", b.ttl)

	// wait for conn and setup ticker
	go func() {
		select {
		case <-b.Control.Ctx().Done():
		case <-b.connDirtyCtx.Done():
			go func() {
				log.Trace("ticker open")
				defer log.Trace("ticker close")
				ticker := time.NewTicker(b.ttl / 2)
				defer ticker.Stop()
				for {
					select {
					case <-b.Control.Ctx().Done():
					case <-ticker.C:
						log.Trace("tick")
						b.operationChan <- nil
					}
				}
			}()
		}
	}()

	pending := map[string]operation{}
	var opErr error
LOOP:
	for {
		select {
		case <-b.Control.Ctx().Done():
			break LOOP
		case ingest := <-b.operationChan:
			log.Debugf("operations received: %v", ingest)
			for _, op := range ingest {
				pending[op.key] = op
			}

			select {
			case <-b.connDirtyCtx.Done():
			default:
				log.Infof("connection is not established")
				continue LOOP
			}

			for _, op := range pending {
				log.Tracef("executing %v", op)
				key := b.chroot + "/" + op.key
				switch op.op {
				case opSetTTL:
					opErr = b.conn.Put(key, []byte(op.value), &store.WriteOptions{
						TTL: b.ttl,
					})
				case opSet:
					opErr = b.conn.Put(key, []byte(op.value), nil)
				case opDelete:
					opErr = b.conn.Delete(key)
					switch opErr {
					case store.ErrKeyNotFound:
						opErr = nil
					}
				}
				if opErr != nil {
					log.Error(opErr)
				} else {
					log.Tracef("op %v executed successfully", op)
					// remove finite op
					switch op.op {
					case opDelete, opSet:
						delete(pending, op.key)
						log.Tracef("finite op %v removed from pending", op)
					}
				}
			}
		}
	}
}

func (b *LibKVBackend) watchLoop(prefix string, consumer bus.Consumer) {
	<-b.connDirtyCtx.Done()
	if b.connErr != nil {
		b.log.Errorf("%v: disabling consumer", b.connErr)
		consumer.ConsumeMessage(bus.NewMessage(prefix, map[string]string{}))
		return
	}
	chroot := b.chroot + "/" + prefix
	log := b.log.GetLog(b.log.Prefix(), "watch", chroot)

	var lastHash uint64 = ^uint64(0)
	cache := map[string]string{}

	collectChan := make(chan []*store.KVPair)
	sleepChan := make(chan struct{})
	var receiving bool
	var retryCount int

	log.Info("open")
LOOP:
	for {
		if !receiving {
			go func() {
				log.Debugf("watching (retry %d)", retryCount)
				changesChan, err := b.conn.WatchTree(chroot, nil)
				if err != nil {
					log.Errorf("failed: %v", err)
				} else {
					log.Tracef("established (retry %d)", retryCount)
					for pairs := range changesChan {
						collectChan <- pairs
					}
					log.Trace("lost")
				}
				select {
				case <-b.Ctx().Done():
					return
				case sleepChan <- struct{}{}:
				}
			}()
		}

		select {
		case <-b.Control.Ctx().Done():
			break LOOP
		case pairs := <-collectChan:
			receiving = true
			log.Tracef("%d pairs received", len(pairs))
			retryCount = 0

			data := map[string]string{}
			for _, pair := range pairs {
				key := strings.TrimPrefix(pair.Key, chroot+"/")
				if pair.Value != nil {
					data[key] = string(pair.Value)
				} else if old, ok := cache[key]; ok {
					log.Tracef("old value is founded for %s:nil", key)
					data[key] = old
				}
			}
			newHash, _ := hashstructure.Hash(data, nil)
			if newHash == lastHash {
				log.Tracef("skipping update: data is equal")
				continue LOOP
			}
			lastHash = newHash
			cache = data
			consumer.ConsumeMessage(bus.NewMessage(prefix, data))
			log.Debugf("consumer updated with %v", data)

		case <-sleepChan:
			log.Trace("sleep request received")
			retryCount++
			receiving = false
			log.Infof("sleeping %v (retry %d)", b.options.RetryInterval, retryCount)
			select {
			case <-b.Ctx().Done():
				break LOOP
			case <-time.After(b.options.RetryInterval):
				log.Debugf("revoking after (retry %d)", retryCount)
				continue LOOP
			}
		}
	}
}
