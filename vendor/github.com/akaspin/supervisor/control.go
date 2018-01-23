package supervisor

import (
	"context"
	"sync/atomic"
)

// Control is simplest embeddable component
type Control struct {
	ctx    context.Context
	cancel context.CancelFunc

	isOpen uint32
}

// NewControl returns new Control
func NewControl(ctx context.Context) (c *Control) {
	c = &Control{}
	c.ctx, c.cancel = context.WithCancel(ctx)
	return
}

func (c *Control) Open() (err error) {
	atomic.CompareAndSwapUint32(&c.isOpen, 0, 1)
	return nil
}

// Close closes Control context
func (c *Control) Close() (err error) {
	c.cancel()
	return nil
}

// Wait blocks until internal done context is not done. See Done().
func (c *Control) Wait() (err error) {
	<-c.ctx.Done()
	return nil
}

// Ctx returns Control context
func (c *Control) Ctx() context.Context {
	return c.ctx
}

func (c *Control) IsOpen() (ok bool) {
	return atomic.LoadUint32(&c.isOpen) == 1
}

func (c *Control) IsClosed() (ok bool) {
	select {
	case <-c.ctx.Done():
		return true
	default:
		return false
	}
}
