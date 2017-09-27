package backend

import (
	"bytes"
	"encoding/json"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
)

// NodeAnnouncer exposes Agent properties in kv
type NodeAnnouncer struct {
	log    *logx.Log
	setter bus.Setter
	prefix string
}

func NewNodesAnnouncer(log *logx.Log, setter bus.Setter, prefix string) (j *NodeAnnouncer) {
	j = &NodeAnnouncer{
		log:    log.GetLog("json", prefix),
		setter: setter,
		prefix: prefix,
	}
	return
}

// Accepts data and pipes it to kv upstream
func (r *NodeAnnouncer) ConsumeMessage(message bus.Message) {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(message.GetPayload()); err != nil {
		r.log.Error(err)
		return
	}
	r.setter.Set(map[string]string{
		r.prefix: buf.String(),
	})
}
