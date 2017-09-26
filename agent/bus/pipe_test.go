// +build ide test_unit

package bus_test

import (
	"context"
	"github.com/akaspin/logx"
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

func TestSimplePipe_Sync(t *testing.T) {
	ctx := context.Background()
	log := logx.GetLog("test")

	cons1 := newTestConsumer()
	cons2 := newTestConsumer()

	pipe := bus.NewSimplePipe(func(message bus.Message) (res bus.Message) {
		payload := message.GetPayload()
		delete(payload, "a")
		res = bus.NewMessage(message.GetPrefix(), payload)
		return
	}, cons1, cons2)

	producer := bus.NewFlatMap(ctx, log, true, "test", pipe)
	assert.NoError(t, producer.Open())

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

	producer.Close()
	producer.Wait()
}
