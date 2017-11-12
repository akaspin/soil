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
	t.Skip()
	consumer := &bus.TestingConsumer{}
	crashChan := make(chan struct{})
	kv := cluster.NewKV(context.Background(), logx.GetLog("test"), cluster.NewTestingBackendFactory(consumer, crashChan, nil))
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

func TestKV_Subscribe(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	consumer := &bus.TestingConsumer{}
	crashChan := make(chan struct{})
	msgChan := make(chan bus.Message, 1)

	kv := cluster.NewKV(ctx, logx.GetLog("test"), cluster.NewTestingBackendFactory(consumer, crashChan, msgChan))
	assert.NoError(t, kv.Open())

	cons1 := &bus.TestingConsumer{}
	ctx1, _ := context.WithCancel(context.Background())
	cons2 := &bus.TestingConsumer{}
	ctx2, cancel2 := context.WithCancel(context.Background())

	t.Run(`subscribe 1`, func(t *testing.T) {
		kv.Subscribe("test/1", ctx1, cons1)
		time.Sleep(time.Millisecond * 200)
	})
	t.Run(`configure 1`, func(t *testing.T) {
		config := cluster.DefaultConfig()
		config.RetryInterval = time.Millisecond * 100
		kv.Configure(config)
		time.Sleep(time.Millisecond * 200)
	})
	t.Run(`put 1`, func(t *testing.T) {
		msgChan <- bus.NewMessage("test/1", map[string]string{"1": "1"})
		time.Sleep(time.Millisecond * 200)
		cons1.AssertMessages(t,
			bus.NewMessage("test/1", map[string]string{"1": "1"}),
		)
	})
	t.Run(`put 1 duplicate`, func(t *testing.T) {
		msgChan <- bus.NewMessage("test/1", map[string]string{"1": "1"})
		time.Sleep(time.Millisecond * 200)
		cons1.AssertMessages(t,
			bus.NewMessage("test/1", map[string]string{"1": "1"}),
		)
	})
	t.Run(`subscribe 2`, func(t *testing.T) {
		kv.Subscribe("test/1", ctx2, cons2)
		time.Sleep(time.Millisecond * 200)
		cons2.AssertMessages(t,
			bus.NewMessage("test/1", map[string]string{"1": "1"}),
		)
	})
	t.Run(`crash`, func(t *testing.T) {
		crashChan <- struct{}{}
		time.Sleep(time.Millisecond * 200)
		cons1.AssertMessages(t,
			bus.NewMessage("test/1", map[string]string{"1": "1"}),
		)
		cons2.AssertMessages(t,
			bus.NewMessage("test/1", map[string]string{"1": "1"}),
		)
	})
	t.Run(`unsubscribe 2`, func(t *testing.T) {
		cancel2()
		time.Sleep(time.Millisecond * 100)
		msgChan <- bus.NewMessage("test/1", map[string]string{"1": "2"})
		time.Sleep(time.Millisecond * 100)
		cons1.AssertMessages(t,
			bus.NewMessage("test/1", map[string]string{"1": "1"}),
			bus.NewMessage("test/1", map[string]string{"1": "2"}),
		)
		cons2.AssertMessages(t,
			bus.NewMessage("test/1", map[string]string{"1": "1"}),
		)
	})
}
