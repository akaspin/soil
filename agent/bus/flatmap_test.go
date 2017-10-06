// +build ide test_unit

package bus_test

import (
	"github.com/akaspin/soil/agent/bus"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

type testConsumer struct {
	mu      sync.Mutex
	records []map[string]string
}

func (c *testConsumer) ConsumeMessage(message bus.Message) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.records = append(c.records, message.GetPayload())
}

func TestFlatMap_Set(t *testing.T) {
	t.Run("strict", func(t *testing.T) {
		cons1 := &testConsumer{}

		prod := bus.NewFlatMap(true, "meta", cons1)
		prod.Set(map[string]string{
			"1": "1",
		})
		assert.Equal(t, cons1.records, []map[string]string{
			{"1": "1"},
		})
		prod.Set(map[string]string{
			"2": "2",
		})
		assert.Equal(t, cons1.records, []map[string]string{
			{"1": "1"},
			{"2": "2"},
		})
		prod.Set(map[string]string{})
		assert.Equal(t, cons1.records, []map[string]string{
			{"1": "1"},
			{"2": "2"},
			{},
		})

	})
	t.Run("non-strict", func(t *testing.T) {
		cons1 := &testConsumer{}

		prod := bus.NewFlatMap(false, "meta", cons1)
		prod.Set(map[string]string{
			"1": "1",
		})
		assert.Equal(t, cons1.records, []map[string]string{
			{"1": "1"},
		})
		prod.Set(map[string]string{
			"2": "2",
		})
		assert.Equal(t, cons1.records, []map[string]string{
			{"1": "1"},
			{"1": "1", "2": "2"},
		})
	})
}
