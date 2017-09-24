// +build ide test_cluster

package kv_test

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/public/kv"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
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

	assert.Equal(t, cons1.res, []map[string]string{{}})
	assert.Equal(t, cons2.res, []map[string]string{{}})
}

func TestBackend_Set(t *testing.T) {
	//t.SkipNow()
	f := newConsulFixture(t)
	defer f.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	backend := kv.NewBackend(ctx, logx.GetLog("test"), kv.Options{
		RetryInterval: time.Millisecond * 200,
		Enabled:       true,
		Timeout:       time.Second,
		URL:           fmt.Sprintf("consul://%s/test", f.Server.HTTPAddr),
		Advertise:     "127.0.0.1:7654",
		TTL:           time.Second * 2,
	})
	err := backend.Open()
	assert.NoError(t, err)

	time.Sleep(time.Millisecond * 200)

	backend.Set(map[string]string{
		"1/pre-ttl": "pre",
	}, true)

	time.Sleep(time.Second)

	cons1 := newDummyConsumer()
	backend.RegisterConsumer("1", cons1)
	time.Sleep(time.Second)

	assert.Equal(t, cons1.res, []map[string]string{
		{"pre-ttl": "pre"},
	})

	backend.Set(map[string]string{
		"1/1": "v1",
		"1/2": "v2",
	}, false)
	time.Sleep(time.Second)

	assert.Equal(t, cons1.res[len(cons1.res)-1], map[string]string{
		"pre-ttl": "pre",
		"1":       "v1",
		"2":       "v2",
	})

	backend.Set(map[string]string{
		"1/ttl1": "ttl1",
		"1/ttl2": "ttl2",
	}, true)
	time.Sleep(time.Second)

	assert.Equal(t, cons1.res[len(cons1.res)-1], map[string]string{
		"pre-ttl": "pre",
		"1":       "v1",
		"2":       "v2",
		"ttl1":    "ttl1",
		"ttl2":    "ttl2",
	})

	backend.Delete("1/1", "1/ttl2")
	time.Sleep(time.Second)
	assert.Equal(t, cons1.res[len(cons1.res)-1], map[string]string{
		"pre-ttl": "pre",
		"2":       "v2",
		"ttl1":    "ttl1",
	})

	// delete non-existent
	backend.Delete("1/fake", "1/ttl2")

	time.Sleep(time.Second * 2)

	backend.Close()
	backend.Wait()

	assert.Equal(t, cons1.res[len(cons1.res)-1], map[string]string{
		"pre-ttl": "pre",
		"2":       "v2",
		"ttl1":    "ttl1",
	})

	for _, chunk := range cons1.res {
		_, ok := chunk["pre-ttl"]
		assert.True(t, ok)
	}
}

func TestBackend_RegisterConsumer_TTL(t *testing.T) {
	//t.SkipNow()
	f := newConsulFixture(t)
	defer f.Stop()

	consul.Register()
	lkv, err := libkv.NewStore(
		store.CONSUL,
		[]string{f.Server.HTTPAddr},
		&store.Config{
			ConnectionTimeout: time.Second,
		},
	)
	assert.NoError(t, err)
	defer lkv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	src := kv.NewBackend(ctx, logx.GetLog("test"), kv.Options{
		RetryInterval: time.Millisecond * 200,
		Enabled:       true,
		Timeout:       time.Second,
		URL:           fmt.Sprintf("consul://%s/test", f.Server.HTTPAddr),
		Advertise:     "127.0.0.1:7654",
		TTL:           time.Second * 2,
	})
	err = src.Open()
	assert.NoError(t, err)

	cons1 := newDummyConsumer()
	src.RegisterConsumer("1", cons1)
	time.Sleep(time.Second)

	// set permanent value
	err = lkv.Put("test/1/permanent", []byte("value"), nil)
	assert.NoError(t, err)

	// set ttl 1s
	base := 1
	err = lkv.Put("test/1/ttl", []byte("val1"), &store.WriteOptions{
		TTL: time.Second * time.Duration(base),
	})
	assert.NoError(t, err)

	time.Sleep(time.Millisecond * 300)
	err = lkv.Put("test/1/ttl", []byte("val2"), &store.WriteOptions{
		TTL: time.Second * time.Duration(base),
	})
	assert.NoError(t, err)

	// wait for expiration
	time.Sleep(time.Second * 2)

	src.Close()
	src.Wait()

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
	lkv, err := libkv.NewStore(
		store.CONSUL,
		[]string{f.Server.HTTPAddr},
		&store.Config{
			ConnectionTimeout: time.Second,
		},
	)
	assert.NoError(t, err)
	defer lkv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	src := kv.NewBackend(ctx, logx.GetLog("test"), kv.Options{
		RetryInterval: time.Millisecond * 200,
		Enabled:       true,
		Timeout:       time.Second,
		URL:           fmt.Sprintf("consul://%s/test", f.Server.HTTPAddr),
		Advertise:     "127.0.0.1:7654",
		TTL:           time.Second * 2,
	})
	err = src.Open()
	assert.NoError(t, err)

	cons1 := newDummyConsumer()
	cons2 := newDummyConsumer()

	src.RegisterConsumer("1", cons1)
	src.RegisterConsumer("2", cons2)

	err = lkv.Put("test/1/1", []byte("1/1-1"), nil)
	err = lkv.Put("test/1/2", []byte("1/2-1"), nil)

	f.Restart(time.Millisecond*500, time.Millisecond*200)

	err = lkv.Put("test/1/2", []byte("1/2-2"), nil)

	time.Sleep(time.Millisecond * 500)

	src.Close()
	src.Wait()

	// cons1
	assert.Equal(t, cons1.res, []map[string]string{
		{}, // established
		{"1": "1/1-1"},
		{"1": "1/1-1", "2": "1/2-1"},
		{"1": "1/1-1", "2": "1/2-2"}, // put
	})

	// cons2
	assert.Equal(t, cons2.res, []map[string]string{
		{},
	})
}

func TestBackend_RegisterConsumer_LateInit(t *testing.T) {
	//t.SkipNow()
	f := newConsulFixture(t)
	defer f.Stop()
	addr := f.Server.HTTPAddr

	consul.Register()
	lkv, err := libkv.NewStore(
		store.CONSUL,
		[]string{addr},
		&store.Config{
			ConnectionTimeout: time.Second,
		},
	)
	assert.NoError(t, err)
	defer lkv.Close()

	// put one record and stop server
	err = lkv.Put("test/1/1", []byte("val"), nil)
	assert.NoError(t, err)
	f.Stop()

	// create backend and bind consumer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	src := kv.NewBackend(ctx, logx.GetLog("test"), kv.Options{
		RetryInterval: time.Millisecond * 200,
		Enabled:       true,
		Timeout:       time.Second,
		URL:           fmt.Sprintf("consul://%s/test", addr),
		Advertise:     "127.0.0.1:7654",
		TTL:           time.Second * 2,
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

	assert.Equal(t, cons1.res, []map[string]string{
		{
			"1": "val",
		},
	})
}
