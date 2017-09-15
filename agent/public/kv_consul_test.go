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
	"github.com/akaspin/supervisor"
	"github.com/davecgh/go-spew/spew"
)

func TestUpdater_Declare_TTL5s(t *testing.T) {
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
		Retry:         5,
		TTL: time.Second * 3,
	})
	updater := public.NewUpdater(ctx, src, "1")
	sv := supervisor.NewChain(ctx, src, updater)
	err = sv.Open()
	assert.NoError(t, err)

	// declare two keys
	updater.Declare(map[string]string{
		"test1": "1",
		"test2": "2",
	})

	time.Sleep(time.Second * 2)

	cons1 := newDummyConsumer()
	src.RegisterConsumer("1", cons1)

	time.Sleep(time.Second * 10)

	sv.Close()
	sv.Wait()

	spew.Dump(cons1.res)
}

func TestKVBackend_RegisterConsumer_TTL(t *testing.T) {
	t.SkipNow()
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
		Retry:         5,
	})
	err = src.Open()
	assert.NoError(t, err)

	cons1 := newDummyConsumer()
	src.RegisterConsumer("1", cons1)
	time.Sleep(time.Second)

	err = kv.Put("test/1/permanent", []byte("value"), nil)
	assert.NoError(t, err)

	base := 2
	err = kv.Put("test/1/ttl", []byte("val1"), nil)
	assert.NoError(t, err)

	time.Sleep(time.Second)
	err = kv.Put("test/1/ttl", []byte("val2"), &store.WriteOptions{
		TTL: time.Second * time.Duration(base),
	})
	assert.NoError(t, err)

	time.Sleep(time.Second * time.Duration(base+2))

	src.Close()
	src.Wait()

	assert.Equal(t, cons1.states, []bool{false, true, true, true, true, true})
	assert.Equal(t, cons1.res, []map[string]string{
		{}, // init
		{}, // established
		{"permanent": "value"},                // perm
		{"permanent": "value", "ttl": "val1"}, // ttl1
		{"permanent": "value", "ttl": "val2"}, // ttl2
		{"permanent": "value"},                // ttl out
	})
}

func TestKVBackend_RegisterConsumer_Recover(t *testing.T) {
	t.SkipNow()
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
		Retry:         5,
	})
	err = src.Open()
	assert.NoError(t, err)

	cons1 := newDummyConsumer()
	cons2 := newDummyConsumer()

	src.RegisterConsumer("1", cons1)
	src.RegisterConsumer("2", cons2)

	err = kv.Put("test/1/1", []byte("1/1-1"), nil)
	err = kv.Put("test/1/2", []byte("1/2-1"), nil)

	f.Restart(time.Second, 0)

	err = kv.Put("test/1/2", []byte("1/2-2"), nil)

	time.Sleep(time.Second)

	// cons1
	assert.Equal(t, cons1.states, []bool{
		false, // init
		true,  // established
		true,
		true,
		false, // lost
		true,  //recovered
		true,  // put
	})
	assert.Equal(t, cons1.res, []map[string]string{
		{}, // init
		{}, // established
		{"1": "1/1-1"},
		{"1": "1/1-1", "2": "1/2-1"},
		{}, // lost
		{"1": "1/1-1", "2": "1/2-1"}, //recovered
		{"1": "1/1-1", "2": "1/2-2"}, // put
	})

	// cons2
	assert.Equal(t, cons2.states, []bool{
		false, // init
		true,  // established
		false, // lost
		true,  // recovered
	})
	assert.Equal(t, cons2.res, []map[string]string{
		{},
		{},
		{},
		{},
	})

	// now stop server and wait for disable
	f.Stop()
	time.Sleep(time.Second * 2)

	// cons1
	assert.Equal(t, cons1.states, []bool{
		false, // init
		true,  // established
		true,
		true,
		false, // lost
		true,  //recovered
		true,
		false, //lost
		true,  //disabled
	})
	assert.Equal(t, cons1.res, []map[string]string{
		{}, // init
		{}, // established
		{"1": "1/1-1"},
		{"1": "1/1-1", "2": "1/2-1"},
		{}, // lost
		{"1": "1/1-1", "2": "1/2-1"}, //recovered
		{"1": "1/1-1", "2": "1/2-2"},
		{}, // lost
		{}, // disabled
	})

	// cons2
	assert.Equal(t, cons2.states, []bool{
		false, // init
		true,  // established
		false, // lost
		true,  // recovered
		false, // lost
		true,  // disabled
	})
	assert.Equal(t, cons2.res, []map[string]string{
		{},
		{},
		{},
		{},
		{},
		{},
	})

	src.Close()
	src.Wait()
}
