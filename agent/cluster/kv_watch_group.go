package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/nu7hatch/gouuid"
)

type watchGroup struct {
	ctx    context.Context
	cancel context.CancelFunc
	log    *logx.Log
	key    string

	cache        bus.Message
	requests     map[string]watcher
	registerChan chan watcher
	outChan      chan string
	resultsChan  chan WatchResult
}

func newWatchGroup(ctx context.Context, log *logx.Log, key string) (g *watchGroup) {
	g = &watchGroup{
		log:          log.GetLog("cluster", "kv", "watch", key),
		key:          key,
		cache:        bus.NewMessage("", nil),
		requests:     map[string]watcher{},
		registerChan: make(chan watcher),
		outChan:      make(chan string),
		resultsChan:  make(chan WatchResult),
	}
	g.ctx, g.cancel = context.WithCancel(ctx)
	go g.loop()
	return
}

func (g *watchGroup) register(req watcher) {
	select {
	case <-g.ctx.Done():
	case g.registerChan <- req:
	}
}

func (g *watchGroup) AcceptResult(result WatchResult) {
	select {
	case <-g.ctx.Done():
	case g.resultsChan <- result:
		g.log.Tracef(`consumed: %v`, result)
	}
}

func (g *watchGroup) loop() {
	log := g.log.GetLog(g.log.Prefix(), append(g.log.Tags(), "loop")...)
LOOP:
	for {
		select {
		case <-g.ctx.Done():
			break LOOP
		case req := <-g.registerChan:
			id, _ := uuid.NewV4()
			g.requests[id.String()] = req
			go func() {
				select {
				case <-g.ctx.Done():
				case <-req.Ctx.Done():
					log.Tracef(`watcher %s: closed`, id.String())
					select {
					case <-g.ctx.Done():
					case g.outChan <- id.String():
						log.Debugf(`cleanup request sent for watcher %s`, id.String())
					}
				}
			}()
			g.log.Tracef("registered %s", id.String())
			if g.cache.GetID() != "" {
				req.consumer.ConsumeMessage(g.cache)
			}
		case id := <-g.outChan:
			delete(g.requests, id)
			if len(g.requests) == 0 {
				g.cancel()
			}
		case result := <-g.resultsChan:
			payload := map[string]interface{}{}
			for k, raw := range result.Data {
				var value interface{}
				if err := json.NewDecoder(bytes.NewReader(raw)).Decode(&value); err != nil {
					g.log.Debug(err)
					continue
				}
				payload[k] = value
			}
			ingest := bus.NewMessage(result.Key, payload)

			if g.cache.Payload().Hash() == ingest.Payload().Hash() {
				g.log.Tracef(`skip broadcast: message is equal to cache`)
				continue LOOP
			}
			g.cache = bus.NewMessage(ingest.GetID(), ingest.Payload())
			g.log.Tracef(`broadcasting: %v to %d consumers`, result, len(g.requests))
			for id, w := range g.requests {
				select {
				case <-w.Ctx.Done():
				default:
					w.consumer.ConsumeMessage(ingest)
					g.log.Tracef(`sent to consumer %s: %v`, id, result)
				}
			}
		}
	}
	log.Debug("done")
}

type watcher struct {
	WatchRequest
	consumer bus.Consumer
}
