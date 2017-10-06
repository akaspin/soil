// +build ide test_unit

package bus_test

import (
	"github.com/akaspin/soil/agent/bus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSimplePipe_ConsumeMessage(t *testing.T) {
	cons1 := &testConsumer{}
	cons2 := &testConsumer{}

	pipe := bus.NewSimplePipe(func(message bus.Message) (res bus.Message) {
		payload := message.GetPayload()
		delete(payload, "a")
		res = bus.NewMessage(message.GetProducer(), payload)
		return
	}, cons1, cons2)

	producer := bus.NewFlatMap(true, "test", pipe)

	time.Sleep(time.Millisecond * 100)

	producer.Set(map[string]string{
		"a": "1",
		"b": "2",
	})
	time.Sleep(time.Millisecond * 100)

	assert.Equal(t, cons1.records, []map[string]string{
		{"b": "2"},
	})
	assert.Equal(t, cons2.records, []map[string]string{
		{"b": "2"},
	})
}
