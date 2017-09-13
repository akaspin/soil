// +build ide test_unit

package source_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/source"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

type dummyConsumer struct {
	changes int32
}

func (c *dummyConsumer) Sync(producer string, active bool, data map[string]string) {
	if active {
		atomic.AddInt32(&c.changes, 1)
	}
}

func TestMapMetadata(t *testing.T) {

	a := source.NewPlain(context.Background(), logx.GetLog("test"), "meta", true)
	a.Open()
	cons := &dummyConsumer{}

	a.RegisterConsumer("test", cons)
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, int32(0), atomic.LoadInt32(&cons.changes))

	a.Configure(map[string]string{
		"first":  "1",
		"second": "3",
	})

	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, int32(1), atomic.LoadInt32(&cons.changes))

	a.Configure(map[string]string{
		"first": "2",
	})
	time.Sleep(time.Millisecond * 300)
	assert.Equal(t, int32(2), atomic.LoadInt32(&cons.changes))

	a.Close()
	a.Wait()
}
