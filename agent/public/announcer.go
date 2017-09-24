package public

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/metadata"
)

// NodeAnnouncer exposes Agent properties in kv
type NodeAnnouncer struct {
	log    *logx.Log
	setter Setter
	prefix string
}

func NewNodesAnnouncer(ctx context.Context, log *logx.Log, setter Setter, prefix string) (j *NodeAnnouncer) {
	j = &NodeAnnouncer{
		log:    log.GetLog("json", prefix),
		setter: setter,
		prefix: prefix,
	}
	return
}

// Accepts data and pipes it to kv upstream
func (r *NodeAnnouncer) ConsumeMessage(message metadata.Message) {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(message.GetPayload()); err != nil {
		r.log.Error(err)
		return
	}
	r.setter.Set(map[string]string{
		r.prefix: buf.String(),
	}, true)
}
