package bus

// TeePipe replicates consumed message to given downstream
type TeePipe struct {
	downstreams []Consumer
}

func NewTeePipe(downstreams ...Consumer) (p *TeePipe) {
	p = &TeePipe{
		downstreams: downstreams,
	}
	return
}

func (p *TeePipe) ConsumerName() string {
	return "tee"
}

func (p *TeePipe) ConsumeMessage(message Message) {
	for _, downstream := range p.downstreams {
		downstream.ConsumeMessage(message)
	}
}
