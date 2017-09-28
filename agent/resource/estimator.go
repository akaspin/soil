package resource

import (
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/bus"
)

type Estimator struct {
}

func (e *Estimator) Configure(all ...Config) {

}

// Initialise Estimator with recovered resources
func (e *Estimator) Init(recovered map[string][]*allocation.Resource) {

}


func (e *Estimator) ConsumeMessage(message bus.Message) {

}
