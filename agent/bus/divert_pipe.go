package bus

import "sync"

// DivertPipe propagates Messages to downstream Consumer until in Divert mode.
// When entering the mode pipe sends predefined message to downstream. Pipe
// sends last consumed message when exit drain mode.
type DivertPipe struct {
	name     string
	consumer Consumer
	inDrain  Message

	mu          sync.Mutex
	last        Message
	isDiverting bool
}

func NewDivertPipe(consumer Consumer, divert Message) (p *DivertPipe) {
	p = &DivertPipe{
		consumer: consumer,
		inDrain:  divert,
	}
	return
}

func (p *DivertPipe) ConsumeMessage(message Message) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.last = message
	if !p.isDiverting {
		p.consumer.ConsumeMessage(p.last)
	}
}

func (p *DivertPipe) Divert(on bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.isDiverting == on {
		return
	}
	p.isDiverting = on
	if p.isDiverting {
		p.consumer.ConsumeMessage(p.inDrain)
	} else {
		p.consumer.ConsumeMessage(p.last)
	}
}
