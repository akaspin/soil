package supervisor

import (
	"context"
	"sync"
	"time"
	"errors"
)

var (
	CloseTimeoutExceeded = errors.New("close timeout exceeded")
)

// Control provides ability to turn any type to supervisor component
type Control struct {
	ctx    context.Context

	// Cancel cancels control context
	Cancel context.CancelFunc

	closeTimeout time.Duration
	boundedWg *sync.WaitGroup

	closeCtx context.Context
	closeCancel context.CancelFunc
}

func NewControl(ctx context.Context) (c *Control) {
	c = NewControlTimeout(ctx, 0)
	return
}

func NewControlTimeout(ctx context.Context, timeout time.Duration) (c *Control) {
	c = &Control{
		closeTimeout: timeout,
		boundedWg: &sync.WaitGroup{},
	}
	c.ctx, c.Cancel = context.WithCancel(ctx)
	c.closeCtx, c.closeCancel = context.WithCancel(context.Background())
	return
}

func (c *Control) Open() (err error) {
	go func() {
		<-c.ctx.Done()
		c.boundedWg.Wait()
		c.closeCancel()
	}()

	return
}

func (c *Control) Close() (err error) {
	c.Cancel()
	return
}

func (c *Control) Wait() (err error) {
	var timeoutChan <-chan time.Time
	if c.closeTimeout > 0 {
		timer := time.NewTimer(c.closeTimeout)
		defer timer.Stop()
		timeoutChan = timer.C
	}
	select {
	case <-c.closeCtx.Done():
	case <-timeoutChan:
		err = CloseTimeoutExceeded
	}
	return
}

// Ctx returns Control context
func (c *Control) Ctx() context.Context {
	return c.ctx
}

// Acquire increases internal lock counter
func (c *Control) Acquire() {
	c.boundedWg.Add(1)
}

func (c *Control) Release() {
	c.boundedWg.Done()
}

