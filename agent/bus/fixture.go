package bus

import (
	"sync"
	"testing"
	"github.com/stretchr/testify/assert"
)

type TestConsumer struct {
	mu sync.Mutex
	messages []Message
}

func (c *TestConsumer) ConsumeMessage(message Message) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messages = append(c.messages, message)
}

func (c *TestConsumer) AssertPayloads(t *testing.T, expect []map[string]string) {
	t.Helper()
	var res []map[string]string
	for _, message := range c.messages {
		res = append(res, message.GetPayload())
	}
	assert.Equal(t, expect, res)
}
