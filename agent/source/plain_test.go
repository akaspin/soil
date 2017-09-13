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
	atomic.AddInt32(&c.changes, 1)
}

func TestMapMetadata(t *testing.T) {

	a := source.NewPlain(context.Background(), logx.GetLog("test"), "meta", true)
	a.Open()
	cons := &dummyConsumer{}

	a.Set(map[string]string{
		"first":  "1",
		"second": "2",
	}, true)
	a.RegisterConsumer("test", cons)
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, int32(0), atomic.LoadInt32(&cons.changes))

	a.Set(map[string]string{
		"first":  "1",
		"second": "3",
	}, true)

	//a.Notify("test", map[string]string{
	//	"${meta.first}": "1",
	//})
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, int32(1), atomic.LoadInt32(&cons.changes))

	a.Set(map[string]string{
		"first": "2",
	}, false)
	time.Sleep(time.Millisecond * 300)
	assert.Equal(t, int32(2), atomic.LoadInt32(&cons.changes))

	a.Close()
	a.Wait()
}
