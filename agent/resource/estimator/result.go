package estimator

import (
	"encoding/json"
	"fmt"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/manifest"
)

// Estimator result
type Result struct {
	Uuid    string // Estimator id
	Message bus.Message
}

// Create new estimator message with "__values"
func NewEstimatorMessage(id string, err error, values manifest.FlatMap) (res bus.Message) {
	if err != nil {
		res = bus.NewMessage(id, manifest.FlatMap{
			"allocated": "false",
			"failure":   fmt.Sprint(err),
		})
		return
	}
	payload := manifest.FlatMap{
		"allocated": "true",
	}
	for k, v := range values {
		payload[k] = v
	}
	buf, _ := json.Marshal(payload)
	payload["__values"] = string(buf)
	res = bus.NewMessage(id, payload)
	return
}
