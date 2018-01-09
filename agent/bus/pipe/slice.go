package pipe

import (
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"sort"
)

// Slicer pipe accepts map[string]interface{} messages and converts them to []interface{}
type Slice struct {
	log      *logx.Log
	consumer bus.Consumer
}

func NewSlice(log *logx.Log, consumer bus.Consumer) (p *Slice) {
	p = &Slice{
		log:      log.GetLog("pipe", "slicer"),
		consumer: consumer,
	}
	return
}

func (p *Slice) ConsumeMessage(message bus.Message) (err error) {
	var v map[string]interface{}
	var keys []string
	var res []interface{}
	if err = message.Payload().Unmarshal(&v); err != nil {
		p.log.Errorf(`can't unmarshal %s to map[string]interface{}: %v`, message, err)
		return
	}
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		res = append(res, v[k])
	}
	p.consumer.ConsumeMessage(bus.NewMessage(message.GetID(), res))
	return
}
