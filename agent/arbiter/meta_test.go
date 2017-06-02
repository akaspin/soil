package arbiter_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/arbiter"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

func TestMetaArbiter(t *testing.T) {
	a := arbiter.NewMapArbiter(context.Background(), logx.GetLog("test"), "meta", true)
	a.Open()
	var changes int32
	callback := func(v map[string]string) {
		atomic.AddInt32(&changes, 1)
	}

	a.Configure(map[string]string{
		"first": "1",
		"second": "2",
	})
	a.RegisterManager(callback)
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, int32(0), atomic.LoadInt32(&changes))

	a.SubmitPod("test", map[string]string{
		"${meta.first}": "1",
	})
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, int32(1), atomic.LoadInt32(&changes))

	a.Configure(map[string]string{
		"first": "2",
		"second": "2",
	})
	time.Sleep(time.Millisecond * 300)
	assert.Equal(t, int32(2), atomic.LoadInt32(&changes))

	a.Close()
	a.Wait()
}
