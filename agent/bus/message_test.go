// +build ide test_unit

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
	assert.True(t, message.IsEmpty())
}

func TestMessage_IsSimple(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		message := bus.NewMessage("test", map[string]string{
			"": "1",
		})
		assert.True(t, message.IsSimple())
	})
	t.Run("one key", func(t *testing.T) {
		message := bus.NewMessage("test", map[string]string{
			"1": "1",
		})
		assert.False(t, message.IsSimple())
	})
	t.Run("two keys", func(t *testing.T) {
		message := bus.NewMessage("test", map[string]string{
			"1": "1",
			"2": "1",
		})
		assert.False(t, message.IsSimple())
	})
}

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

func TestMessage_Expand(t *testing.T) {
	t.Run("usual", func(t *testing.T) {
		msg := bus.NewMessage("test", map[string]string{
			"1": "1",
			"2": "1",
		})
		assert.Equal(t, map[string]string{
			"test.1": "1",
			"test.2": "1",
		}, msg.Expand())
	})
	t.Run("empty", func(t *testing.T) {
		msg := bus.NewMessage("test", map[string]string{
			"1": "1",
			"2": "1",
		})
		assert.Equal(t, map[string]string{
			"test.1": "1",
			"test.2": "1",
		}, msg.Expand())
	})
}
