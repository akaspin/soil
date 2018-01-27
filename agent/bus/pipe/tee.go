package pipe

import "github.com/akaspin/soil/agent/bus"

// Tee replicates consumed message to given downstreams
type Tee struct {
	downstreams []bus.Consumer
}

func NewTee(downstreams ...bus.Consumer) (p *Tee) {
	return &Tee{
		downstreams: downstreams,
	}
}

func (p *Tee) ConsumeMessage(message bus.Message) (err error) {
	for _, downstream := range p.downstreams {
		downstream.ConsumeMessage(message)
	}
	return nil
}
