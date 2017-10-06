// +build ide test_unit

package bus_test

import (
	"github.com/akaspin/soil/agent/bus"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

type testConsumer struct {
	mu       *sync.Mutex
	messages []bus.Message
}

func newTestConsumer() (c *testConsumer) {
	c = &testConsumer{
		mu: &sync.Mutex{},
	}
	return
}

func (c *testConsumer) ConsumeMessage(message bus.Message) {
	go func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.messages = append(c.messages, message)
	}()
}

func TestSimplePipe_ConsumeMessage(t *testing.T) {
	cons1 := newTestConsumer()
	cons2 := newTestConsumer()

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

	assert.Equal(t, cons1.messages, []bus.Message{
		bus.NewMessage("test", map[string]string{"b": "2"}),
	})
	assert.Equal(t, cons2.messages, []bus.Message{
		bus.NewMessage("test", map[string]string{"b": "2"}),
	})
}
