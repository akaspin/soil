package public

import (
	"github.com/akaspin/supervisor"
	"github.com/akaspin/logx"
	"context"
	"time"
	"github.com/docker/libkv/store"
)

type Updater struct {
	*supervisor.Control
	log *logx.Log

	backend *KVBackend
	prefix string

	declared map[string]string
	declareChan chan map[string]string
}

func NewUpdater(ctx context.Context, backend *KVBackend, prefix string) (u *Updater) {
	u = &Updater{
		Control: supervisor.NewControl(ctx),
		log: backend.log.GetLog(backend.log.Prefix(), append(backend.log.Tags(), []string{"updater", prefix}...)...),
		prefix: prefix,
		backend: backend,
		declared: map[string]string{},
		declareChan: make(chan map[string]string, 1),
	}
	return
}

func (u *Updater) Open() (err error) {
	go u.declareLoop()
	err = u.Control.Open()
	return
}

// Declare data. Data will be updated with TTL
func (u *Updater) Declare(data map[string]string) {
	select {
	case <-u.Control.Ctx().Done():
		return
	default:
	}
	u.declareChan<- data
}

func (u *Updater) declareLoop() {
	u.Acquire()
	defer u.Release()

	// wait for conn
	<-u.backend.dirtyCtx.Done()
	if u.backend.failure != nil {
		u.log.Infof("connection is not established: %v", u.backend.failure)
		return
	}


	u.log.Infof("updating")
	chroot := u.backend.chroot + "/" + u.prefix

	var putErr error
	actualiseChan := make(chan []string, 1)
	ticker := time.NewTicker(u.backend.options.TTL / 2)
	defer ticker.Stop()
LOOP:
	for {
		select {
		case <-u.Control.Ctx().Done():
			break LOOP
		case chunk := <-u.declareChan:
			if len(chunk) == 0 {
				continue LOOP
			}
			var keys []string
			for k, v := range chunk {
				keys = append(keys, k)
				u.declared[k] = v
			}
			actualiseChan <- keys
		case keys := <-actualiseChan:
			u.log.Debugf("storing (restrict %v)", keys)
			for k, v := range u.declared {
				putErr = u.backend.kv.Put(chroot + "/" + k, []byte(v), &store.WriteOptions{
					TTL: u.backend.options.TTL,
				})
				if putErr != nil {
					u.log.Error(putErr)
					continue LOOP
				}
				u.log.Debugf("stored %s", k)
			}
		case <-ticker.C:
			actualiseChan <- nil
		}
	}
}

// Remove keys from KV
func (u *Updater) Remove(keys ...string) {

}