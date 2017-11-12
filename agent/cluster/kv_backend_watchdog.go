package cluster

import (
	"github.com/akaspin/logx"
	"time"
)

// Backend watchdog evaluates Backend contexts and commit channel
type kvBackendWatchdog struct {
	kv      *KV
	backend Backend
	config  Config
	log     *logx.Log
}

func newBackendWatchdog(kv *KV, backend Backend, config Config) (w *kvBackendWatchdog) {
	w = &kvBackendWatchdog{
		kv:      kv,
		backend: backend,
		config:  config,
		log:     kv.log.GetLog("cluster", "watchdog", config.URL, config.ID),
	}
	go w.ready()
	go w.done()
	go w.downstream()
	return
}

// watch ready context
func (w *kvBackendWatchdog) ready() {
	select {
	case <-w.backend.Ctx().Done():
		return
	case <-w.backend.ReadyCtx().Done():
		w.log.Trace(`backend is ready`)
		select {
		case <-w.kv.Control.Ctx().Done():
		case w.kv.invokePendingChan <- struct{}{}:
			w.log.Debug(`try request sent`)
		}
	}
}

func (w *kvBackendWatchdog) done() {
	<-w.backend.Ctx().Done()
	select {
	case <-w.backend.FailCtx().Done():
		w.log.Tracef(`failed: sending wake request after %s`, w.config.RetryInterval)
		select {
		case <-w.kv.Control.Ctx().Done():
		case <-time.After(w.config.RetryInterval):
			select {
			case <-w.kv.Control.Ctx().Done():
			case w.kv.configRequestChan <- kvConfigRequest{
				config:   w.config,
				internal: true,
			}:
				w.log.Trace(`wake request sent`)
			}
		}
	default:
		w.log.Trace(`context canceled`)
	}
}

func (w *kvBackendWatchdog) downstream() {
	w.log.Trace(`downstream: open`)
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
		case msg := <-w.backend.WatchChan():
			w.log.Tracef(`received watch: %v`, msg)
			select {
			case <-w.backend.Ctx().Done():
				w.log.Tracef(`skip sending watch %v: backend closed`)
				break LOOP
			case w.kv.watchResultsChan <- msg:
				w.log.Tracef(`watch sent: %v`, msg)
			}
		}
	}
	w.log.Trace(`downstream: close`)
}
