package cluster

import (
	"github.com/akaspin/logx"
	"time"
)

// Backend watchdog evaluates Backend contexts and commit channel
type kvWatchdog struct {
	kv      *KV
	backend Backend
	config  Config
	log     *logx.Log
}

func newWatchdog(kv *KV, backend Backend, config Config) (w *kvWatchdog) {
	w = &kvWatchdog{
		kv:      kv,
		backend: backend,
		config:  config,
		log:     kv.log.GetLog("cluster", "kv", "watchdog", config.BackendURL, config.NodeID),
	}
	go w.ready()
	go w.done()
	go w.downstream()
	return w
}

// watch ready context
func (w *kvWatchdog) ready() {
	w.log.Trace(`ready: start`)
	select {
	case <-w.backend.Ctx().Done():
		return //
	case <-w.backend.ReadyCtx().Done():
		w.log.Info(`backend is ready`)
		select {
		case <-w.backend.Ctx().Done():
		case w.kv.invokePendingChan <- struct{}{}:
			w.log.Debug(`try request sent`)
		}
	}
}

func (w *kvWatchdog) done() {
	w.log.Trace(`done: start`)
	<-w.backend.Ctx().Done()
	w.log.Trace(`done: backend closed`)
	select {
	case <-w.backend.FailCtx().Done():
		w.log.Errorf(`backend failed: sending wake request after %s`, w.config.RetryInterval)
		select {
		case <-w.kv.Control.Ctx().Done():
			w.log.Trace(`skip wake: kv closed`)
		case <-time.After(w.config.RetryInterval):
			w.log.Trace(`sending reconfigure request`)
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
		w.log.Info(`backend closed`)
	}
}

func (w *kvWatchdog) downstream() {
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
		case results := <-w.backend.WatchResultsChan():
			w.log.Tracef(`received watch: %v`, results)
			select {
			case <-w.backend.Ctx().Done():
				w.log.Tracef(`skip sending watch watch result %v: backend closed`)
				break LOOP
			case w.kv.watchResultsChan <- results:
				w.log.Tracef(`watch result sent: %v`, results)
			}
		}
	}
	w.log.Trace(`downstream: close`)
}
