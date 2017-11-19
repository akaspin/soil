package bus

import (
	"github.com/akaspin/logx"
	"sort"
)

// Slicer pipe accepts map[string]interface{} messages and converts them to []interface{}
type SlicerPipe struct {
	log      *logx.Log
	consumer Consumer
}

func NewSlicerPipe(log *logx.Log, consumer Consumer) (p *SlicerPipe) {
	p = &SlicerPipe{
		log:      log.GetLog("pipe", "slicer"),
		consumer: consumer,
	}
	return
}

func (p *SlicerPipe) ConsumeMessage(message Message) {
	var v map[string]interface{}
	var keys []string
	var res []interface{}
	if err := message.Payload().Unmarshal(&v); err != nil {
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
	p.consumer.ConsumeMessage(NewMessage(message.GetID(), res))
}
