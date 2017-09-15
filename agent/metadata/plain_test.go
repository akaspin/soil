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

	a := metadata.NewPlain(context.Background(), logx.GetLog("test"), "meta", false)
	a.Open()
	cons1 := &dummyConsumer{}
	cons2 := &dummyConsumer{}

	a.RegisterConsumer("test1", cons1)
	a.RegisterConsumer("test2", cons2)

	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, int32(0), atomic.LoadInt32(&cons1.changes))
	assert.Equal(t, int32(0), atomic.LoadInt32(&cons2.changes))

	a.Configure(map[string]string{
		"first":  "1",
		"second": "3",
	})

	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, int32(1), atomic.LoadInt32(&cons1.changes))
	assert.Equal(t, int32(1), atomic.LoadInt32(&cons2.changes))

	a.Configure(map[string]string{
		"first": "2",
	})
	time.Sleep(time.Millisecond * 300)
	assert.Equal(t, int32(2), atomic.LoadInt32(&cons1.changes))
	assert.Equal(t, int32(2), atomic.LoadInt32(&cons2.changes))

	a.Close()
	a.Wait()
}
