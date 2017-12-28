package bus

type MutatePipe struct {
	Downstream Consumer
	Fn         func(message Message) Message
}

func NewFnPipe(fn func(message Message) Message, downstream Consumer) (p *MutatePipe) {
	p = &MutatePipe{
		Fn:         fn,
		Downstream: downstream,
	}
	return
}

func (p *MutatePipe) ConsumeMessage(message Message) (err error) {
	err = p.Downstream.ConsumeMessage(p.Fn(message))
	return
}
