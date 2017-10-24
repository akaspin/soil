package bus

type SimplePipe struct {
	consumers []Consumer
	fn        func(message Message) Message
}

func NewSimplePipe(fn func(message Message) Message, consumers ...Consumer) (p *SimplePipe) {
	p = &SimplePipe{
		fn:        fn,
		consumers: consumers,
	}
	return
}

func (p *SimplePipe) ConsumeMessage(message Message) {
	res := message
	if p.fn != nil {
		res = p.fn(res)
	}
	for _, consumer := range p.consumers {
		consumer.ConsumeMessage(res)
	}
}
