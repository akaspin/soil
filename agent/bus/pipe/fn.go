package pipe

import "github.com/akaspin/soil/agent/bus"

type Fn struct {
	Downstream bus.Consumer
	Fn         func(message bus.Message) bus.Message
}

func NewFn(fn func(message bus.Message) bus.Message, downstream bus.Consumer) (p *Fn) {
	p = &Fn{
		Fn:         fn,
		Downstream: downstream,
	}
	return
}

func (p *Fn) GetConsumer() (c bus.Consumer) {
	return p.Downstream
}

func (p *Fn) ConsumeMessage(message bus.Message) (err error) {
	err = p.Downstream.ConsumeMessage(p.Fn(message))
	return
}
