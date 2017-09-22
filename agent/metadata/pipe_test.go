// +build ide test_unit

package metadata_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/metadata"
	"github.com/akaspin/supervisor"
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

func (c *testConsumer) Sync(message metadata.Message) {
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
	}, cons1.Sync, cons2.Sync)

	producer := metadata.NewSimpleProducer(ctx, log, "test", pipe.Sync)
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

func TestPipe_Sync(t *testing.T) {
	ctx := context.Background()
	log := logx.GetLog("test")

	cons1 := newTestConsumer()
	producer := metadata.NewSimpleProducer(ctx, log, "test")
	pipe := metadata.NewPipe(ctx, log, "pipe", producer, func(message metadata.Message) (res metadata.Message) {
		delete(message.Data, "a")
		res = message
		return
	}, cons1.Sync)
	sv := supervisor.NewChain(ctx, producer, pipe)
	assert.NoError(t, sv.Open())

	cons2 := newTestConsumer()
	//go pipe.RegisterConsumer("1", cons1)
	go pipe.RegisterConsumer("2", cons2.Sync)

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

	sv.Close()
	sv.Wait()
}
