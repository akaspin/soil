package estimator

import (
	"fmt"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/manifest"
)

// Create new estimator message
func NewEstimatorMessage(id string, err error, values manifest.Environment) (res bus.Message) {
	if err != nil {
		res = bus.NewMessage(id, manifest.Environment{
			"allocated": "false",
			"failure":   fmt.Sprint(err),
		})
		return
	}
	payload := manifest.Environment{
		"allocated": "true",
	}
	for k, v := range values {
		payload[k] = v
	}
	res = bus.NewMessage(id, payload)
	return
}
