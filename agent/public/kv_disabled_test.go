// +build ide test_cluster

package public_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/public"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestKVBackend_RegisterConsumer_Disabled(t *testing.T) {
	t.SkipNow()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	src := public.NewKVBackend(ctx, logx.GetLog("test"), public.BackendOptions{
		RetryInterval: time.Millisecond * 300,
		Enabled:       false,
		Timeout:       time.Second,
		URL:           "",
		Advertise:     "127.0.0.1:7654",
		Retry:         0,
	})
	err := src.Open()
	assert.NoError(t, err)

	cons1 := newDummyConsumer()
	cons2 := newDummyConsumer()

	src.RegisterConsumer("1", cons1)
	src.RegisterConsumer("2", cons2)

	time.Sleep(time.Second)

	src.Close()
	src.Wait()

	assert.Equal(t, cons1.states, []bool{false, true})
	assert.Equal(t, cons1.res, []map[string]string{{}, {}})
	assert.Equal(t, cons2.states, []bool{false, true})
	assert.Equal(t, cons2.res, []map[string]string{{}, {}})
}
