// +build ide test_unit

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

	a := source.NewMap(context.Background(), logx.GetLog("test"), "meta", true, manifest.Constraint{})
	a.Open()
	var changes int32
	callback := func(active bool, v map[string]string) {
		atomic.AddInt32(&changes, 1)
	}

	a.Set(map[string]string{
		"first":  "1",
		"second": "2",
	}, true)
	a.Register(callback)
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, int32(0), atomic.LoadInt32(&changes))

	a.SubmitPod("test", map[string]string{
		"${meta.first}": "1",
	})
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, int32(1), atomic.LoadInt32(&changes))

	a.Set(map[string]string{
		"first": "2",
	}, false)
	time.Sleep(time.Millisecond * 300)
	assert.Equal(t, int32(2), atomic.LoadInt32(&changes))

	a.Close()
	a.Wait()
}
