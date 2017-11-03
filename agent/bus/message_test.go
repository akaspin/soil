// build ide test_unit

package bus_test

import (
	"fmt"
	"github.com/akaspin/soil/agent/bus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMessageStringify(t *testing.T) {
	var res interface{}
	res = "a"
	assert.Equal(t, "a", fmt.Sprint(res))

	res = 1
	assert.Equal(t, "1", fmt.Sprint(res))

	res = true
	assert.Equal(t, "true", fmt.Sprint(res))

	res = 127.9
	assert.Equal(t, "127.9", fmt.Sprint(res))
}

func TestMessage_IsEmpty(t *testing.T) {
	message := bus.NewMessage("test", nil)
	assert.True(t, message.Payload().IsEmpty())
}

func TestMessage_GetPayload(t *testing.T) {
	payload := map[string]string{
		"1": "1",
	}
	msg := bus.NewMessage("test", payload)
	payload["2"] = "2"

	var chunk map[string]string
	assert.NoError(t, msg.Payload().Unmarshal(&chunk))
	assert.NotEqual(t, chunk, payload)
	assert.Equal(t, chunk, map[string]string{
		"1": "1",
	})
}
