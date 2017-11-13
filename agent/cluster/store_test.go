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

func TestStore_ConsumeMessage(t *testing.T) {
	consumer := &bus.TestingConsumer{}
	crashChan := make(chan struct{})
	kv := cluster.NewKV(context.Background(), logx.GetLog("test"), cluster.NewTestingBackendFactory(consumer, crashChan, nil))
	assert.NoError(t, kv.Open())

	t.Run(`configure`, func(t *testing.T) {
		config := cluster.DefaultConfig()
		config.RetryInterval = time.Millisecond * 10
		kv.Configure(config)
		time.Sleep(time.Millisecond * 100)
	})
	t.Run(`volatile with prefix`, func(t *testing.T) {
		store := cluster.NewVolatileStore(kv, "prefix")
		store.ConsumeMessage(bus.NewMessage("1", map[string]string{
			"1": "1",
		}))
		time.Sleep(time.Millisecond * 100)
		consumer.AssertMessages(t,
			bus.NewMessage("test", map[string]interface{}{
				"prefix/1": map[string]interface{}{
					"Data": map[string]string{"1": "1"},
					"TTL":  true,
				},
			}),
		)
		store.ConsumeMessage(bus.NewMessage("", map[string]string{
			"1": "2",
		}))
		time.Sleep(time.Millisecond * 100)
		consumer.AssertMessages(t,
			bus.NewMessage("test", map[string]interface{}{
				"prefix/1": map[string]interface{}{
					"Data": map[string]string{"1": "1"},
					"TTL":  true,
				},
			}),
			bus.NewMessage("test", map[string]interface{}{
				"prefix": map[string]interface{}{
					"Data": map[string]string{"1": "2"},
					"TTL":  true,
				},
			}),
		)
	})
	t.Run(`volatile without prefix`, func(t *testing.T) {
		store := cluster.NewVolatileStore(kv, "")
		store.ConsumeMessage(bus.NewMessage("1", map[string]string{
			"1": "1",
		}))
		time.Sleep(time.Millisecond * 100)
		consumer.AssertMessages(t,
			bus.NewMessage("test", map[string]interface{}{
				"prefix/1": map[string]interface{}{
					"Data": map[string]string{"1": "1"},
					"TTL":  true,
				},
			}),
			bus.NewMessage("test", map[string]interface{}{
				"prefix": map[string]interface{}{
					"Data": map[string]string{"1": "2"},
					"TTL":  true,
				},
			}),
			bus.NewMessage("test", map[string]interface{}{
				"1": map[string]interface{}{
					"Data": map[string]string{"1": "1"},
					"TTL":  true,
				},
			}),
		)
	})

}
