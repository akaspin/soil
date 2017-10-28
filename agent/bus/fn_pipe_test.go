// +build ide test_unit

package bus_test

import (
	"github.com/akaspin/soil/agent/bus"
	"testing"
	"time"
)

func TestFnPipe_ConsumeMessage(t *testing.T) {
	c1 := &bus.DummyConsumer{}
	c2 := &bus.DummyConsumer{}

	pipe := bus.NewFnPipe(func(message bus.Message) (res bus.Message) {
		payload := message.GetPayloadMap()
		delete(payload, "a")
		res = bus.NewMessage(message.GetID(), payload)
		return
	}, c1, c2)

	pipe.ConsumeMessage(bus.NewMessage("test", map[string]string{
		"a": "1",
		"b": "2",
	}))
	time.Sleep(time.Millisecond * 100)

	time.Sleep(time.Millisecond * 100)

	c1.AssertPayloads(t, []map[string]string{
		{"b": "2"},
	})
	c2.AssertPayloads(t, []map[string]string{
		{"b": "2"},
	})
}
