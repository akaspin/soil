// +build ide test_cluster

package kv_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/public/kv"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestKVBackend_RegisterConsumer_Disabled(t *testing.T) {
	//t.SkipNow()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	src := kv.NewBackend(ctx, logx.GetLog("test"), kv.Options{
		RetryInterval: time.Millisecond * 300,
		Enabled:       false,
		Timeout:       time.Second,
		URL:           "",
		Advertise:     "127.0.0.1:7654",
	})
	err := src.Open()
	assert.NoError(t, err)

	cons1 := newDummyConsumer()
	cons2 := newDummyConsumer()

	src.RegisterConsumer("1", cons1)
	src.RegisterConsumer("2", cons2)

	time.Sleep(time.Millisecond * 200)

	src.Close()
	src.Wait()

	assert.Equal(t, cons1.states, []bool{true})
	assert.Equal(t, cons1.res, []map[string]string{{}})
	assert.Equal(t, cons2.states, []bool{true})
	assert.Equal(t, cons2.res, []map[string]string{{}})
}
