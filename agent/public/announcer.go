package public

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/metadata"
	"github.com/akaspin/soil/agent/public/kv"
)

// NodeAnnouncer exposes Agent properties in kv
type NodeAnnouncer struct {
	log    *logx.Log
	setter kv.Setter
	prefix string
}

func NewNodesAnnouncer(ctx context.Context, log *logx.Log, backend kv.Setter, prefix string) (j *NodeAnnouncer) {
	j = &NodeAnnouncer{
		log:    log.GetLog("json", prefix),
		setter: backend,
		prefix: prefix,
	}
	return
}

// Accepts data and pipes it to kv upstream
func (r *NodeAnnouncer) Sync(message metadata.Message) {
	if !message.Clean {
		return
	}
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(message.Data); err != nil {
		r.log.Error(err)
		return
	}
	r.setter.Set(map[string]string{
		r.prefix: buf.String(),
	}, true)
}
