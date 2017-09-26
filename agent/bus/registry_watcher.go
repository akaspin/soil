package bus

import (
	"encoding/json"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/manifest"
	"strings"
)

type PublicRegistryWatcher struct {
	log *logx.Log
	consumers []RegistryConsumer
}

func NewWatcher(log *logx.Log, consumers ...RegistryConsumer) (w *PublicRegistryWatcher) {
	w = &PublicRegistryWatcher{
		log: log.GetLog("public", "watch", "registry"),
		consumers: consumers,
	}
	return
}

func (w *PublicRegistryWatcher) ConsumeMessage(message Message) {
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



