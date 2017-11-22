package bus

type FnPipe struct {
	consumers []Consumer
	fn        func(message Message) Message
}

func NewFnPipe(fn func(message Message) Message, consumers ...Consumer) (p *FnPipe) {
	p = &FnPipe{
		fn:        fn,
		consumers: consumers,
	}
	return
}

func (p *FnPipe) ConsumeMessage(message Message) (err error) {
	res := message
	if p.fn != nil {
		res = p.fn(res)
	}
	for _, consumer := range p.consumers {
		consumer.ConsumeMessage(res)
	}
	return
}
