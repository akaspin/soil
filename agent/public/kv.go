package public

import (
	"context"
	"errors"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/supervisor"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"github.com/docker/libkv/store/zookeeper"
	"github.com/mitchellh/hashstructure"
	"net/url"
	"strings"
	"time"
	"github.com/akaspin/soil/agent/metadata"
)

var disabledError = errors.New("public namespace is disabled")

type BackendOptions struct {
	Enabled       bool
	Advertise     string
	URL           string
	Timeout       time.Duration
	Retry         int
	RetryInterval time.Duration
	TTL           time.Duration
}

func (o BackendOptions) ParseUrl() (kind store.Backend, chroot string, addr []string, err error) {
	if !o.Enabled {
		err = disabledError
		return
	}
	u, err := url.Parse(o.URL)
	if err != nil {
		return
	}
	kind = store.Backend(u.Scheme)
	chroot = strings.TrimPrefix(u.Path, "/")
	addr = strings.Split(u.Host, ",")

	switch kind {
	case store.CONSUL:
		consul.Register()
	case store.ETCD:
		etcd.Register()
	case store.ZK:
		zookeeper.Register()
	default:
		err = fmt.Errorf("invalid backend type: %s", kind)
	}
	return
}

type KVBackend struct {
	*supervisor.Control
	log     *logx.Log
	options BackendOptions

	// chroot cache
	kind   store.Backend
	chroot string
	addr   []string

	kv          store.Store
	failure     error
	dirtyCtx    context.Context
	dirtyCancel context.CancelFunc
}

func NewKVBackend(ctx context.Context, log *logx.Log, options BackendOptions) (b *KVBackend) {
	b = &KVBackend{
		Control: supervisor.NewControl(ctx),
		options: options,
	}
	b.kind, b.chroot, b.addr, b.failure = b.options.ParseUrl()
	b.dirtyCtx, b.dirtyCancel = context.WithCancel(context.Background())
	b.log = log.GetLog("kv", b.chroot)
	return
}

func (b *KVBackend) Open() (err error) {
	go b.connect()
	err = b.Control.Open()
	return
}

// Registers Consumer with specific prefix
func (b *KVBackend) RegisterConsumer(prefix string, consumer metadata.Consumer) {

	chroot := b.chroot + "/" + prefix
	log := b.log.GetLog(b.log.Prefix(), append(b.log.Tags(), []string{"watch", prefix}...)...)

	go func() {
		b.Acquire()
		defer func() {
			log.Debug("close")
			b.Release()
		}()
		b.deactivateConsumer(prefix, consumer)

		// wait for conn
		<-b.dirtyCtx.Done()
		if b.failure != nil {
			log.Infof("connection is not established: %v: disabling", b.failure)
			b.disableConsumer(prefix, consumer)
			return
		}

		log.Infof("watching")

		// permanent collect chan
		collectChan := make(chan []*store.KVPair)
		watchdogChan := make(chan struct{})
		sleepChan := make(chan struct{})

		// send initial watchdog
		go func() {
			log.Debug("sending initial watchdog request")
			select {
			case <-b.Ctx().Done():
				return
			default:
			}
			watchdogChan <- struct{}{}
		}()

		var lastHash uint64 = ^uint64(0)
		cache := map[string]string{}
		var retry int

		for {
			select {
			case <-b.Control.Ctx().Done():
				log.Debugf("context is done")
				return
			case pairs := <-collectChan:
				log.Debugf("%d pairs received", len(pairs))
				retry = 0
				data := map[string]string{}
				for _, pair := range pairs {
					key := strings.TrimPrefix(pair.Key, chroot+"/")
					if pair.Value != nil {
						data[key] = string(pair.Value)
					} else if old, ok := cache[key]; ok {
						log.Debugf("old value is founded for %s:nil", key)
						data[key] = old
					}
				}
				if h, _ := hashstructure.Hash(data, nil); h != lastHash {
					lastHash = h
					cache = data
					consumer.Sync(metadata.Message{
						Prefix: prefix,
						Clean: true,
						Data: data,
					})
					log.Debugf("consumer updated with %v", data)
				} else {
					log.Debugf("skipping update: data is equal")
				}
			case <-sleepChan:
				log.Debug("sleep request received")
				retry++
				if b.options.Retry > 0 && retry > b.options.Retry {
					log.Errorf("%d retries exceed: disabling", b.options.Retry)
					b.disableConsumer(prefix, consumer)
					return
				}
				log.Infof("retry %d of %d: sleeping %v in inactive state", retry, b.options.Retry, b.options.RetryInterval)
				lastHash = ^uint64(0)
				if retry == 1 {
					b.deactivateConsumer(prefix, consumer)
				}
				go func() {
					select {
					case <-b.Ctx().Done():
						return
					case <-time.After(b.options.RetryInterval):
						select {
						case <-b.Ctx().Done():
							return
						case watchdogChan <- struct{}{}:
						}
					}
				}()
			case <-watchdogChan:
				log.Debug("watchdog request received")
				select {
				case <-b.Ctx().Done():
					return
				default:
				}
				go func() {
					changesChan, err := b.kv.WatchTree(chroot, nil)
					if err != nil {
						log.Errorf("failed: %v", err)
					} else {
						log.Debug("established")
						for pairs := range changesChan {
							collectChan <- pairs
						}
						log.Debug("lost")
					}
					select {
					case <-b.Ctx().Done():
						return
					case sleepChan <- struct{}{}:
					}
				}()
			}
		}
	}()
}

func (b *KVBackend) connect() {
	b.Acquire()
	defer b.Release()
	defer b.dirtyCancel()

	if b.failure != nil {
		b.log.Warning(b.failure)
		return
	}

	var retry int
	for {
		retry++
		b.log.Infof("connecting to %s (retry %d)", b.options.URL, retry)
		b.kv, b.failure = libkv.NewStore(b.kind, b.addr, &store.Config{
			ConnectionTimeout: b.options.Timeout,
		})
		if b.failure != nil {
			b.log.Errorf("failed to connect to %s: %v", b.options.URL, b.failure)
			if b.options.Retry > 0 && retry >= b.options.Retry {
				b.failure = fmt.Errorf("exceed %d retries to connect to %s", b.options.Retry, b.options.URL)
				b.log.Error(b.failure)
				return
			}
			b.log.Infof("sleeping %s before reconnect to %s", b.options.RetryInterval, b.options.URL)
			time.Sleep(b.options.RetryInterval)
			continue
		} else {
			b.log.Infof("connected to %s", b.options.URL)
			return
		}
	}
}



func (b *KVBackend) disableConsumer(prefix string, consumer metadata.Consumer) {
	consumer.Sync(metadata.Message{
		Prefix: prefix,
		Clean: true,
		Data: map[string]string{},
	})
}

func (b *KVBackend) deactivateConsumer(prefix string, consumer metadata.Consumer) {
	consumer.Sync(metadata.Message{
		Prefix: prefix,
		Clean: false,
		Data: map[string]string{},
	})
}


