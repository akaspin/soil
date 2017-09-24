// +build ide test_unit

package metadata_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/metadata"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

type testConsumer struct {
	mu       *sync.Mutex
	messages []metadata.Message
}

func newTestConsumer() (c *testConsumer) {
	c = &testConsumer{
		mu: &sync.Mutex{},
	}
	return
}

func (c *testConsumer) ConsumeMessage(message metadata.Message) {
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

	pipe := metadata.NewSimplePipe(func(message metadata.Message) (res metadata.Message) {
		delete(message.Data, "a")
		res = message
		return
	}, cons1, cons2)

	producer := metadata.NewSimpleProducer(ctx, log, "test", pipe)
	assert.NoError(t, producer.Open())

	time.Sleep(time.Millisecond * 100)

	producer.Replace(map[string]string{
		"a": "1",
		"b": "2",
	})
	time.Sleep(time.Millisecond * 100)

	assert.Equal(t, cons1.messages, []metadata.Message{
		{Clean: true, Prefix: "test", Data: map[string]string{"b": "2"}}})
	assert.Equal(t, cons2.messages, []metadata.Message{
		{Clean: true, Prefix: "test", Data: map[string]string{"b": "2"}}})

	producer.Close()
	producer.Wait()
}
