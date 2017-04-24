package concurrency_test

import (
	"context"
	"github.com/akaspin/concurrency"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type nopCloser struct {
	closeCounter *int64
}

func (n *nopCloser) Close() (err error) {
	atomic.AddInt64(n.closeCounter, 1)
	return
}

func TestNewResourcePool(t *testing.T) {
	var factoryCount, closeCount int64
	type resource struct {
		Close func() (err error)
	}

	p := concurrency.NewResourcePool(
		context.TODO(),
		concurrency.Config{
			Capacity:     16,
			CloseTimeout: time.Millisecond * 100,
		},
		func() (r concurrency.Resource, err error) {
			r = &nopCloser{&closeCount}
			atomic.AddInt64(&factoryCount, 1)
			return
		})
	p.Open()

	wg := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := p.Get(context.TODO())
			assert.NoError(t, err)
			defer p.Put(r)
		}()
	}

	wg.Wait()

	assert.Equal(t, factoryCount, int64(16))
	assert.Equal(t, closeCount, int64(0))

	p.Close()
	err := p.Wait()

	assert.NoError(t, err)
	assert.Equal(t, factoryCount, int64(16))
	assert.Equal(t, closeCount, int64(16))

}
