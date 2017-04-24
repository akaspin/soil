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

func TestNewWorkerPool(t *testing.T) {
	var count int64

	fn := func() {
		atomic.AddInt64(&count, 1)
	}

	p := concurrency.NewWorkerPool(context.TODO(), concurrency.Config{
		CloseTimeout: time.Millisecond * 200,
		Capacity:     16,
	})
	err := p.Open()
	assert.NoError(t, err)

	wg := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go p.Execute(context.TODO(), func() {
			defer wg.Done()
			fn()
		})
	}
	wg.Wait()

	err = p.Close()
	assert.NoError(t, err)

	err = p.Wait()
	assert.NoError(t, err)

	assert.EqualValues(t, 100, count)

}

func TestWorkerPool_Wait(t *testing.T) {
	var count int64

	p := concurrency.NewWorkerPool(context.TODO(), concurrency.Config{
		//CloseTimeout: time.Millisecond * 200,
		Capacity: 16,
	})
	err := p.Open()
	assert.NoError(t, err)

	fn := func() {
		time.Sleep(time.Second * 1)
		res1 := atomic.AddInt64(&count, 1)
		if res1 >= 15 {
			p.Close()
		}
	}

	for i := 0; i < 100; i++ {
		go p.Execute(context.TODO(), func() {
			fn()
		})
	}

	//err = p.Close()
	//assert.NoError(t, err)

	err = p.Wait()
	assert.NoError(t, err)

	assert.True(t, count < 100)
	assert.True(t, count > 15)
	t.Log(count)
}
