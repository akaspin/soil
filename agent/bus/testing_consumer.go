package bus

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"fmt"
	"reflect"
)

// Dummy Consumer for testing purposes
type TestingConsumer struct {
	mu       sync.RWMutex
	messages []Message
}

func (c *TestingConsumer) ConsumerName() string {
	return "testing"
}

func (c *TestingConsumer) ConsumeMessage(message Message) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.messages = append(c.messages, message)
}

func (c *TestingConsumer) AssertPayloads(t *testing.T, expect []map[string]string) {
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

func (c *TestingConsumer) Messages() []Message {
	return c.messages
}

func (c *TestingConsumer) CompareMessagesFn(expect ...Message) (fn func() error) {
	fn = func() (err error) {
		var eq bool
		c.mu.RLock()
		eq = reflect.DeepEqual(expect, c.messages)
		c.mu.RUnlock()
		if !eq {
			err = fmt.Errorf("not equal (expected)%s != (actual)%s", expect, c.messages)
		}
		return
	}
	return
}

func (c *TestingConsumer) AssertMessages(t *testing.T, expect ...Message) {
	t.Helper()
	assert.Equal(t, expect, c.messages)
}

func (c *TestingConsumer) AssertLastMessage(t *testing.T, expect Message) {
	t.Helper()
	if len(c.messages) == 0 {
		assert.Equal(t, expect, nil)
		return
	}
	assert.Equal(t, expect, c.messages[len(c.messages)-1])
}
