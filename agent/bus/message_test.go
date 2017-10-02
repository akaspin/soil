// +build ide test_unit

package bus_test

import (
	"github.com/akaspin/soil/agent/bus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMessage_GetPayload(t *testing.T) {
	payload := map[string]string{
		"1": "1",
	}
	msg := bus.NewMessage("test", payload)
	payload["2"] = "2"
	assert.NotEqual(t, msg.GetPayload(), payload)
	assert.Equal(t, msg.GetPayload(), map[string]string{
		"1": "1",
	})
}
