package supervisor_test

import (
	"context"
	"errors"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type component500 struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func newComponent500() (c *component500) {
	c = &component500{}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	return c
}

func (*component500) Open() (err error) {
	return nil
}

func (c *component500) Close() (err error) {
	c.cancel()
	return nil
}

func (c *component500) Wait() (err error) {
	<-c.ctx.Done()
	<-time.After(time.Millisecond * 300) // forether
	return nil
}

func TestTimeout_Wait(t *testing.T) {
	for i := 0; i < compositeTestIterations; i++ {
		t.Run("ok", func(t *testing.T) {
			t.Parallel()
			c1 := newTestingComponent("1", nil, nil, errors.New("1"))
			to := supervisor.NewTimeout(context.Background(), time.Second, c1)
			assert.NoError(t, to.Open())

			assert.NoError(t, to.Close())
			assert.EqualError(t, to.Wait(), "1")
			c1.assertCycle(t)
		})
		t.Run("timeout", func(t *testing.T) {
			t.Parallel()
			c1 := newComponent500()
			to := supervisor.NewTimeout(context.Background(), time.Millisecond*50, c1)
			assert.NoError(t, to.Open())
			assert.NoError(t, to.Close())
			assert.EqualError(t, to.Wait(), supervisor.ErrTimeout.Error())
		})
		t.Run("inside", func(t *testing.T) {
			t.Parallel()
			c1 := newTestingComponent("1", nil, nil, errors.New("1"))
			to := supervisor.NewTimeout(context.Background(), time.Second, c1)
			assert.NoError(t, to.Open())

			close(c1.closedChan)
			assert.EqualError(t, to.Wait(), "1")
			c1.assertEvents(t, "open", "done")
		})
	}
}
