package pipe

import "github.com/akaspin/soil/agent/bus"

// Blackhole pipe consumes messages and silently discards them
type Blackhole struct{}

func (*Blackhole) ConsumeMessage(message bus.Message) (err error) {
	return
}
