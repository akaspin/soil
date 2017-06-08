package source_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/source"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

func TestMapMetadata(t *testing.T) {
	a := source.NewMapSource(context.Background(), logx.GetLog("test"), "meta", true, manifest.Constraint{})
	a.Open()
	var changes int32
	callback := func(active bool, v map[string]string) {
		atomic.AddInt32(&changes, 1)
	}

	a.Configure(map[string]string{
		"first":  "1",
		"second": "2",
	})
	a.Register(callback)
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, int32(0), atomic.LoadInt32(&changes))

	a.SubmitPod("test", map[string]string{
		"${meta.first}": "1",
	})
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, int32(1), atomic.LoadInt32(&changes))

	a.Configure(map[string]string{
		"first":  "2",
		"second": "2",
	})
	time.Sleep(time.Millisecond * 300)
	assert.Equal(t, int32(2), atomic.LoadInt32(&changes))

	a.Close()
	a.Wait()
}
