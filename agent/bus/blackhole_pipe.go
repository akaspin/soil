package bus

type BlackholePipe struct{}

func (*BlackholePipe) ConsumeMessage(message Message) (err error) {
	return
}
