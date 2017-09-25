package registry

import (
	"encoding/json"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/metadata"
	"github.com/akaspin/soil/manifest"
	"strings"
)

type Watcher struct {
	log *logx.Log
	consumers []Consumer
}

func NewWatcher(log *logx.Log, consumers ...Consumer) (w *Watcher) {
	w = &Watcher{
		log: log.GetLog("public", "watch", "registry"),
		consumers: consumers,
	}
	return
}

func (w *Watcher) ConsumeMessage(message metadata.Message) {
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



