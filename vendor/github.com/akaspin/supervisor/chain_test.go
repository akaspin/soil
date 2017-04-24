package supervisor_test

import (
	"context"
	"errors"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
)

type chainableSig struct {
	index int
	op    string
}

type chainable struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup

	index int

	closeCount *int64
	ch         chan chainableSig
	err        error
}

func newChainable(index int, rCh chan chainableSig, closeCount *int64) (c *chainable) {
	c = &chainable{
		index:      index,
		ch:         rCh,
		closeCount: closeCount,
		wg:         &sync.WaitGroup{},
	}
	c.ctx, c.cancel = context.WithCancel(context.TODO())
	return
}

func (c *chainable) Open() (err error) {
	c.wg.Add(1)
	go func() {
		<-c.ctx.Done()
		c.wg.Done()
		c.ch <- chainableSig{c.index, "done"}
	}()
	c.ch <- chainableSig{c.index, "open"}
	return
}

func (c *chainable) Close() (err error) {
	atomic.AddInt64(c.closeCount, 1)
	//println("close", c.index)
	c.cancel()
	return
}

func (c *chainable) Wait() (err error) {
	c.wg.Wait()
	c.ch <- chainableSig{c.index, "wait"}
	err = c.err
	return
}

func (c *chainable) Crash(err error) {
	c.err = err
	c.cancel()
}

func TestChain_Empty(t *testing.T) {
	c := supervisor.NewChain(context.TODO())
	c.Open()
	c.Wait()
}

func TestChain_OK(t *testing.T) {
	var closeCount int64
	resCh := make(chan chainableSig)
	var res []chainableSig
	resWg := &sync.WaitGroup{}
	resWg.Add(1)
	go func() {
		for i := 0; i < 9; i++ {
			res = append(res, <-resCh)
		}
		resWg.Done()
	}()

	c := supervisor.NewChain(
		context.TODO(),
		newChainable(1, resCh, &closeCount),
		newChainable(2, resCh, &closeCount),
		newChainable(3, resCh, &closeCount),
	)
	c.Open()
	c.Close()
	err := c.Wait()
	resWg.Wait()

	assert.NoError(t, err)
	assert.Equal(t, res, []chainableSig{
		{index: 1, op: "open"},
		{index: 2, op: "open"},
		{index: 3, op: "open"},
		{index: 3, op: "done"},
		{index: 3, op: "wait"},
		{index: 2, op: "done"},
		{index: 2, op: "wait"},
		{index: 1, op: "done"},
		{index: 1, op: "wait"}})
	assert.Equal(t, int64(3), closeCount)
}

func TestChain_Crash(t *testing.T) {
	var closeCount int64
	resCh := make(chan chainableSig)
	var res []chainableSig
	resWg := &sync.WaitGroup{}
	resWg.Add(1)
	go func() {
		for i := 0; i < 9; i++ {
			res = append(res, <-resCh)
		}
		resWg.Done()
	}()

	messy := newChainable(1, resCh, &closeCount)
	g := supervisor.NewChain(
		context.TODO(),
		newChainable(2, resCh, &closeCount),
		newChainable(3, resCh, &closeCount),
		messy,
	)
	g.Open()
	messy.Crash(errors.New("err"))
	err := g.Wait()
	resWg.Wait()

	assert.Error(t, err)
	assert.Equal(t, "err", err.Error())
	assert.Equal(t, res, []chainableSig{
		{index: 2, op: "open"},
		{index: 3, op: "open"},
		{index: 1, op: "open"},
		{index: 1, op: "done"},
		{index: 1, op: "wait"},
		{index: 3, op: "done"},
		{index: 3, op: "wait"},
		{index: 2, op: "done"},
		{index: 2, op: "wait"}})
	assert.Equal(t, int64(3), closeCount)
}
