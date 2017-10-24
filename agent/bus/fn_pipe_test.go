// +build ide test_unit

package bus_test

import (
	"github.com/akaspin/soil/agent/bus"
	"testing"
	"time"
)

func TestTeePipe_ConsumeMessage(t *testing.T) {
	c1 := &bus.DummyConsumer{}
	c2 := &bus.DummyConsumer{}

	pipe := bus.NewFnPipe(func(message bus.Message) (res bus.Message) {
		payload := message.GetPayload()
		delete(payload, "a")
		res = bus.NewMessage(message.GetPrefix(), payload)
		return
	}, c1, c2)

	producer := bus.NewStrictMapUpstream("test", pipe)

	time.Sleep(time.Millisecond * 100)

	producer.Set(map[string]string{
		"a": "1",
		"b": "2",
	})
	time.Sleep(time.Millisecond * 100)

	c1.AssertPayloads(t, []map[string]string{
		{"b": "2"},
	})
	c2.AssertPayloads(t, []map[string]string{
		{"b": "2"},
	})
}
