package public

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/metadata"
	"github.com/akaspin/soil/agent/public/kv"
)

// Announcer exposes Agent properties in kv
type Announcer struct {
	backend *kv.Backend
	*metadata.SimpleProducer

	log *logx.Log

	//upstream metadata.Producer
	prefix   string
}

func NewAnnouncer(ctx context.Context, log *logx.Log, backend *kv.Backend, prefix string, upstream metadata.Producer) (j *Announcer) {
	j = &Announcer{
		backend:        backend,
		log:            log.GetLog("json", prefix),
		SimpleProducer: metadata.NewSimpleProducer(ctx, log, prefix),
		prefix:         prefix,
		//upstream:       upstream,
	}
	return
}

func (r *Announcer) Open() (err error) {
	//r.upstream.RegisterConsumer(r.prefix, r.Sync)
	err = r.SimpleProducer.Open()
	return
}

// Accepts data and pipes it to kv upstream
func (r *Announcer) Sync(message metadata.Message) {
	if !message.Clean {
		return
	}
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(message.Data); err != nil {
		r.log.Error(err)
		return
	}
	r.backend.Set(map[string]string{
		r.prefix: buf.String(),
	}, true)
}
