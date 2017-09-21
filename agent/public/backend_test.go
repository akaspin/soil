// +build ide test_cluster

package public_test

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/public"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBackend_RegisterConsumer_TTL(t *testing.T) {
	//t.SkipNow()
	f := newConsulFixture(t)
	defer f.Stop()

	consul.Register()
	kv, err := libkv.NewStore(
		store.CONSUL,
		[]string{f.Server.HTTPAddr},
		&store.Config{
			ConnectionTimeout: time.Second,
		},
	)
	assert.NoError(t, err)
	defer kv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	src := public.NewKVBackend(ctx, logx.GetLog("test"), public.BackendOptions{
		RetryInterval: time.Millisecond * 200,
		Enabled:       true,
		Timeout:       time.Second,
		URL:           fmt.Sprintf("consul://%s/test", f.Server.HTTPAddr),
		Advertise:     "127.0.0.1:7654",
	})
	err = src.Open()
	assert.NoError(t, err)

	cons1 := newDummyConsumer()
	src.RegisterConsumer("1", cons1)
	time.Sleep(time.Second)

	// set permanent value
	err = kv.Put("test/1/permanent", []byte("value"), nil)
	assert.NoError(t, err)

	// set ttl 1s
	base := 1
	err = kv.Put("test/1/ttl", []byte("val1"), &store.WriteOptions{
		TTL: time.Second * time.Duration(base),
	})
	assert.NoError(t, err)

	time.Sleep(time.Millisecond * 300)
	err = kv.Put("test/1/ttl", []byte("val2"), &store.WriteOptions{
		TTL: time.Second * time.Duration(base),
	})
	assert.NoError(t, err)

	// wait for expiration
	time.Sleep(time.Second * 2)

	src.Close()
	src.Wait()

	assert.Equal(t, cons1.states, []bool{
		true,
		true,
		true,
		true,
		true,
	})
	assert.Equal(t, cons1.res, []map[string]string{
		{}, // established
		{"permanent": "value"},                // perm
		{"permanent": "value", "ttl": "val1"}, // ttl1
		{"permanent": "value", "ttl": "val2"}, // ttl2
		{"permanent": "value"},                // ttl out
	})
}

func TestBackend_RegisterConsumer_Recover(t *testing.T) {
	//t.SkipNow()
	f := newConsulFixture(t)
	defer f.Stop()

	consul.Register()
	kv, err := libkv.NewStore(
		store.CONSUL,
		[]string{f.Server.HTTPAddr},
		&store.Config{
			ConnectionTimeout: time.Second,
		},
	)
	assert.NoError(t, err)
	defer kv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	src := public.NewKVBackend(ctx, logx.GetLog("test"), public.BackendOptions{
		RetryInterval: time.Millisecond * 200,
		Enabled:       true,
		Timeout:       time.Second,
		URL:           fmt.Sprintf("consul://%s/test", f.Server.HTTPAddr),
		Advertise:     "127.0.0.1:7654",
	})
	err = src.Open()
	assert.NoError(t, err)

	cons1 := newDummyConsumer()
	cons2 := newDummyConsumer()

	src.RegisterConsumer("1", cons1)
	src.RegisterConsumer("2", cons2)

	err = kv.Put("test/1/1", []byte("1/1-1"), nil)
	err = kv.Put("test/1/2", []byte("1/2-1"), nil)

	f.Restart(time.Millisecond*500, time.Millisecond*200)

	err = kv.Put("test/1/2", []byte("1/2-2"), nil)

	time.Sleep(time.Millisecond * 500)

	src.Close()
	src.Wait()

	// cons1
	assert.Equal(t, cons1.states, []bool{
		true, // established
		true,
		true,
		false, // lost
		true,  //recovered
		true,  // put
	})
	assert.Equal(t, cons1.res, []map[string]string{
		{}, // established
		{"1": "1/1-1"},
		{"1": "1/1-1", "2": "1/2-1"},
		nil, // lost
		{"1": "1/1-1", "2": "1/2-1"}, //recovered
		{"1": "1/1-1", "2": "1/2-2"}, // put
	})

	// cons2
	assert.Equal(t, cons2.states, []bool{
		true,  // established
		false, // lost
		true,  // recovered
	})
	assert.Equal(t, cons2.res, []map[string]string{
		{},
		nil,
		{},
	})
}

func TestBackend_RegisterConsumer_LateInit(t *testing.T) {
	f := newConsulFixture(t)
	defer f.Stop()
	addr := f.Server.HTTPAddr

	consul.Register()
	kv, err := libkv.NewStore(
		store.CONSUL,
		[]string{addr},
		&store.Config{
			ConnectionTimeout: time.Second,
		},
	)
	assert.NoError(t, err)
	defer kv.Close()

	// put one record and stop server
	err = kv.Put("test/1/1", []byte("val"), nil)
	assert.NoError(t, err)
	f.Stop()

	// create backend and bind consumer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	src := public.NewKVBackend(ctx, logx.GetLog("test"), public.BackendOptions{
		RetryInterval: time.Millisecond * 200,
		Enabled:       true,
		Timeout:       time.Second,
		URL:           fmt.Sprintf("consul://%s/test", addr),
		Advertise:     "127.0.0.1:7654",
	})
	err = src.Open()
	assert.NoError(t, err)

	cons1 := newDummyConsumer()
	src.RegisterConsumer("1", cons1)
	time.Sleep(time.Millisecond * 400)

	// start server
	f.Start()
	time.Sleep(time.Millisecond * 400)

	src.Close()
	src.Wait()

	assert.Equal(t, cons1.states, []bool{
		false,
		true,
	})
	assert.Equal(t, cons1.res, []map[string]string{
		nil,
		{
			"1": "val",
		},
	})
}
