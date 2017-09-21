package public

import (
	"context"
	"errors"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/metadata"
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
)

var disabledError = errors.New("public namespace is disabled")

type BackendOptions struct {
	Enabled       bool
	Advertise     string
	URL           string
	Timeout       time.Duration
	RetryInterval time.Duration
	TTL           time.Duration
}

type Backend struct {
	*supervisor.Control
	log     *logx.Log
	options BackendOptions

	// connection context
	connDirtyCtx    context.Context
	connDirtyCancel context.CancelFunc
	conn            store.Store
	connErr         error
	chroot          string
	ttl             time.Duration
}

func NewKVBackend(ctx context.Context, log *logx.Log, options BackendOptions) (b *Backend) {
	b = &Backend{
		Control: supervisor.NewControl(ctx),
		options: options,
	}
	b.connDirtyCtx, b.connDirtyCancel = context.WithCancel(context.Background())
	b.log = log.GetLog("kv")
	return
}

func (b *Backend) Open() (err error) {
	go b.connect()
	err = b.Control.Open()
	return
}

// Registers Consumer with specific prefix
func (b *Backend) RegisterConsumer(prefix string, consumer metadata.Consumer) {
	go func() {
		// wait for conn
		<-b.connDirtyCtx.Done()
		if b.connErr != nil {
			if b.connErr != disabledError {
				b.log.Errorf("%v: disabling consumer", b.connErr)
			}
			consumer.Sync(metadata.Message{
				Prefix: prefix,
				Clean:  true,
				Data:   map[string]string{},
			})
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
	LOOP:
		for {
			if !receiving {
				go func() {
					log.Infof("establishing (retry %d)", retryCount)
					changesChan, err := b.conn.WatchTree(chroot, nil)
					if err != nil {
						log.Errorf("failed: %v", err)
					} else {
						log.Debugf("established (retry %d)", retryCount)
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

			select {
			case <-b.Control.Ctx().Done():
				break LOOP
			case pairs := <-collectChan:
				receiving = true
				log.Debugf("%d pairs received", len(pairs))
				retryCount = 0

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
				newHash, _ := hashstructure.Hash(data, nil)
				if newHash == lastHash {
					log.Debugf("skipping update: data is equal")
					continue LOOP
				}
				lastHash = newHash
				cache = data
				consumer.Sync(metadata.Message{
					Prefix: prefix,
					Clean:  true,
					Data:   data,
				})
				log.Debugf("consumer updated with %v", data)

			case <-sleepChan:
				log.Debug("sleep request received")
				retryCount++
				receiving = false
				lastHash = ^uint64(0)
				if retryCount == 1 {
					consumer.Sync(metadata.Message{
						Prefix: prefix,
						Clean:  false,
					})
				}
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
	}()
}

func (b *Backend) connect() {
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
