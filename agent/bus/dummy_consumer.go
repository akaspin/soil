package bus

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

// Dummy Consumer for testing purposes
type DummyConsumer struct {
	mu       sync.Mutex
	messages []Message
}

func (c *DummyConsumer) ConsumeMessage(message Message) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messages = append(c.messages, message)
}

func (c *DummyConsumer) AssertPayloads(t *testing.T, expect []map[string]string) {
	t.Helper()
	var res []map[string]string
	for _, message := range c.messages {
		var chunk map[string]string
		if err := message.Payload().Unmarshal(&chunk); err != nil {
			t.Error(err)
			t.Fail()
		}
		res = append(res, chunk)
	}
	assert.Equal(t, expect, res)
}

func (c *DummyConsumer) AssertMessages(t *testing.T, expect ...Message) {
	t.Helper()
	assert.Equal(t, expect, c.messages)
}

func (c *DummyConsumer) AssertLastMessage(t *testing.T, expect Message) {
	t.Helper()
	if len(c.messages) == 0 {
		assert.Equal(t, expect, nil)
		return
	}
	assert.Equal(t, expect, c.messages[len(c.messages)-1])
}
