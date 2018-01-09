package pipe

import (
	"github.com/akaspin/soil/agent/bus"
	"sync"
)

// Divert propagates Messages to downstream Consumer until in Divert mode.
// When entering in divert mode pipe sends predefined message to downstream.
// Pipe sends last consumed message when exit divert mode.
type Divert struct {
	name     string
	consumer bus.Consumer
	inDrain  bus.Message

	mu          sync.Mutex
	last        bus.Message
	isDiverting bool
}


// Create new divert pipe with given consumer and divert message
func NewDivert(consumer bus.Consumer, divert bus.Message) (d *Divert) {
	d = &Divert{
		consumer: consumer,
		inDrain:  divert,
	}
	return
}

func (d *Divert) GetConsumer() (c bus.Consumer) {
	c = d.consumer
	return
}

// Consume message from upstream and resend it to downstream then not in divert mode. Otherwise send predefined divert message.
func (d *Divert) ConsumeMessage(message bus.Message) (err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.last = message
	if !d.isDiverting {
		d.consumer.ConsumeMessage(d.last)
	}
	return
}

// Divert sets Divert pipe state
func (d *Divert) Divert(on bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.isDiverting == on {
		return
	}
	d.isDiverting = on
	if d.isDiverting {
		d.consumer.ConsumeMessage(d.inDrain)
	} else {
		d.consumer.ConsumeMessage(d.last)
	}
}
