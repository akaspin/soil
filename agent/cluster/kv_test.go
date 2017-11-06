// +build ide test_unit

package cluster_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/cluster"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestKV_Configure(t *testing.T) {
	consumer := &bus.TestingConsumer{}
	crashChan := make(chan struct{})
	kv := cluster.NewKV(context.Background(), logx.GetLog("test"), cluster.NewTestingBackendFactory(consumer, crashChan))
	assert.NoError(t, kv.Open())

	t.Run(`submit on zero`, func(t *testing.T) {
		kv.Submit([]cluster.BackendStoreOp{
			{bus.NewMessage("pre-volatile", map[string]string{"1": "1"}), true},
			{bus.NewMessage("pre-permanent", map[string]string{"1": "1"}), false},
		})
		time.Sleep(time.Millisecond * 300)
		consumer.AssertMessages(t)
	})
	t.Run(`configure 1`, func(t *testing.T) {
		config := cluster.DefaultConfig()
		config.RetryInterval = time.Millisecond * 100
		kv.Configure(config)
		time.Sleep(time.Millisecond * 300)

	})
	t.Run(`ensure submit after config`, func(t *testing.T) {
		consumer.AssertMessages(t,
			bus.NewMessage("test", map[string]interface{}{
				"pre-volatile": map[string]interface{}{
					"Data": map[string]string{"1": "1"},
					"TTL":  true,
				},
				"pre-permanent": map[string]interface{}{
					"Data": map[string]string{"1": "1"},
					"TTL":  false,
				},
			}),
		)
	})
	t.Run(`ensure resubmit volatile after crash`, func(t *testing.T) {
		crashChan <- struct{}{}
		time.Sleep(time.Millisecond * 300)

		consumer.AssertMessages(t,
			bus.NewMessage("test", map[string]interface{}{
				"pre-volatile": map[string]interface{}{
					"Data": map[string]string{"1": "1"},
					"TTL":  true,
				},
				"pre-permanent": map[string]interface{}{
					"Data": map[string]string{"1": "1"},
					"TTL":  false,
				},
			}),
			bus.NewMessage("test", map[string]interface{}{
				"pre-volatile": map[string]interface{}{
					"Data": map[string]string{"1": "1"},
					"TTL":  true,
				},
			}),
		)
	})
	t.Run(`ensure noop after equal config`, func(t *testing.T) {
		config := cluster.DefaultConfig()
		config.RetryInterval = time.Millisecond * 100
		kv.Configure(config)
		time.Sleep(time.Millisecond * 300)
		consumer.AssertMessages(t,
			bus.NewMessage("test", map[string]interface{}{
				"pre-volatile": map[string]interface{}{
					"Data": map[string]string{"1": "1"},
					"TTL":  true,
				},
				"pre-permanent": map[string]interface{}{
					"Data": map[string]string{"1": "1"},
					"TTL":  false,
				},
			}),
			bus.NewMessage("test", map[string]interface{}{
				"pre-volatile": map[string]interface{}{
					"Data": map[string]string{"1": "1"},
					"TTL":  true,
				},
			}),
		)
	})
	t.Run(`add and remove`, func(t *testing.T) {
		kv.Submit([]cluster.BackendStoreOp{
			{bus.NewMessage("pre-volatile", nil), true},
			{bus.NewMessage("post-volatile", map[string]string{"1": "1"}), true},
		})
		time.Sleep(time.Millisecond * 300)
		consumer.AssertMessages(t,
			bus.NewMessage("test", map[string]interface{}{
				"pre-volatile": map[string]interface{}{
					"Data": map[string]string{"1": "1"},
					"TTL":  true,
				},
				"pre-permanent": map[string]interface{}{
					"Data": map[string]string{"1": "1"},
					"TTL":  false,
				},
			}),
			bus.NewMessage("test", map[string]interface{}{
				"pre-volatile": map[string]interface{}{
					"Data": map[string]string{"1": "1"},
					"TTL":  true,
				},
			}),
			bus.NewMessage("test", map[string]interface{}{
				"pre-volatile": map[string]interface{}{
					"Data": nil,
					"TTL":  true,
				},
				"post-volatile": map[string]interface{}{
					"Data": map[string]string{"1": "1"},
					"TTL":  true,
				},
			}),
		)
	})

	kv.Close()
	kv.Wait()
}
