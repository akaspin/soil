// +build ide test_unit

package bus_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

type dummyConsumer struct {
	changes int32
}

func (c *dummyConsumer) ConsumeMessage(message bus.Message) {
	atomic.AddInt32(&c.changes, 1)
}

func TestMapMetadata(t *testing.T) {

	cons1 := &dummyConsumer{}
	cons2 := &dummyConsumer{}
	a := bus.NewFlatMap(context.Background(), logx.GetLog("test"), true, "meta",
		cons1, cons2)
	a.Open()

	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, int32(0), atomic.LoadInt32(&cons1.changes))
	assert.Equal(t, int32(0), atomic.LoadInt32(&cons2.changes))

	a.Set(map[string]string{
		"first":  "1",
		"second": "3",
	})

	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, int32(1), atomic.LoadInt32(&cons1.changes))
	assert.Equal(t, int32(1), atomic.LoadInt32(&cons2.changes))

	a.Set(map[string]string{
		"first": "2",
	})
	time.Sleep(time.Millisecond * 300)
	assert.Equal(t, int32(2), atomic.LoadInt32(&cons1.changes))
	assert.Equal(t, int32(2), atomic.LoadInt32(&cons2.changes))

	a.Close()
	a.Wait()
}
