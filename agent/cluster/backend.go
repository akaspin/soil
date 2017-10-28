package cluster

import (
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/supervisor"
)

//
type Backend struct {
	*supervisor.Control
	log *logx.Log

	worker Worker
	config *Config

	state map[string]bus.Message

	storeTTLChan chan bus.Message
	storeChan    chan bus.Message
	configChan   chan Config
}

func (b *Backend) Configure(config Config) {

}
