// +build ide test_unit

package metadata_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/metadata"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

type dummyConsumer struct {
	changes int32
}

func (c *dummyConsumer) Sync(message metadata.Message) {
	if message.Clean {
		atomic.AddInt32(&c.changes, 1)
	}
}

func TestMapMetadata(t *testing.T) {

	cons1 := &dummyConsumer{}
	cons2 := &dummyConsumer{}
	a := metadata.NewSimpleProducer(context.Background(), logx.GetLog("test"), "meta",
		cons1.Sync, cons2.Sync)
	a.Open()

	//a.RegisterConsumer("test1", cons1.Sync)
	//a.RegisterConsumer("test2", cons2.Sync)

	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, int32(0), atomic.LoadInt32(&cons1.changes))
	assert.Equal(t, int32(0), atomic.LoadInt32(&cons2.changes))

	a.Replace(map[string]string{
		"first":  "1",
		"second": "3",
	})

	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, int32(1), atomic.LoadInt32(&cons1.changes))
	assert.Equal(t, int32(1), atomic.LoadInt32(&cons2.changes))

	a.Replace(map[string]string{
		"first": "2",
	})
	time.Sleep(time.Millisecond * 300)
	assert.Equal(t, int32(2), atomic.LoadInt32(&cons1.changes))
	assert.Equal(t, int32(2), atomic.LoadInt32(&cons2.changes))

	a.Close()
	a.Wait()
}
