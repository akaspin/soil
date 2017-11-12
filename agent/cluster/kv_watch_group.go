package cluster

import (
	"context"
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
	messageChan  chan bus.Message
}

func newWatchGroup(ctx context.Context, log *logx.Log, key string) (g *watchGroup) {
	g = &watchGroup{
		log:          log.GetLog("cluster", "watch", key),
		key:          key,
		cache:        bus.NewMessage("", nil),
		requests:     map[string]watcher{},
		registerChan: make(chan watcher),
		outChan:      make(chan string),
		messageChan:  make(chan bus.Message, 100),
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

func (g *watchGroup) ConsumeMessage(message bus.Message) {
	select {
	case <-g.ctx.Done():
	case g.messageChan <- message:
		g.log.Tracef(`consumed: %v`, message)
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
		case msg := <-g.messageChan:
			if g.cache.Payload().Hash() == msg.Payload().Hash() {
				g.log.Tracef(`skip broadcast: message is equal to cache`)
				continue LOOP
			}
			g.cache = bus.NewMessage(msg.GetID(), msg.Payload())
			g.log.Tracef(`broadcasting: %v to %d consumers`, msg, len(g.requests))
			for id, w := range g.requests {
				select {
				case <-w.Ctx.Done():
				default:
					w.consumer.ConsumeMessage(msg)
					g.log.Tracef(`sent to consumer %s: %v`, id, msg)
				}
			}
		}
	}
	log.Debug("done")
}

type watcher struct {
	BackendWatchRequest
	consumer bus.Consumer
}
