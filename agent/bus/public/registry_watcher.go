package public

import (
	"encoding/json"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/manifest"
	"strings"
)

type RegistryPipe struct {
	log       *logx.Log
	consumers []scheduler.RegistryConsumer
}

func NewRegistryWatcher(log *logx.Log, consumers ...scheduler.RegistryConsumer) (w *RegistryPipe) {
	w = &RegistryPipe{
		log:       log.GetLog("public", "watch", "registry"),
		consumers: consumers,
	}
	return
}

func (w *RegistryPipe) ConsumeMessage(message bus.Message) {
	var res manifest.Registry
	for _, raw := range message.GetPayload() {
		var pod manifest.Pod
		if err := json.NewDecoder(strings.NewReader(raw)).Decode(&pod); err != nil {
			w.log.Error(err)
			continue
		}
		res = append(res, &pod)
	}
	for _, consumer := range w.consumers {
		consumer.ConsumeRegistry("public", res)
	}
}
