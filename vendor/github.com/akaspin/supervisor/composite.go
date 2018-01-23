package supervisor

import (
	"context"
	"sync"
	"sync/atomic"
)

type compositeControl struct {
	ctx        context.Context
	cancelFunc context.CancelFunc

	open uint32

	closeWg sync.WaitGroup
	waitWg  sync.WaitGroup // WG to wait for exit of all components

	openError  compositeError
	closeError compositeError
	waitError  compositeError // composite Wait() error
}

func (c *compositeControl) isOpen() (ok bool) {
	return atomic.LoadUint32(&c.open) == 1
}

func (c *compositeControl) tryOpen() (op bool) {
	return atomic.CompareAndSwapUint32(&c.open, 0, 1)
}

type composite struct {
	handler func(control *compositeControl)
	control *compositeControl
}

func newComposite(ctx context.Context, handler func(control *compositeControl)) (c *composite) {
	c = &composite{
		handler: handler,
		control: &compositeControl{},
	}
	c.control.ctx, c.control.cancelFunc = context.WithCancel(ctx)
	return c
}

// Open blocks until all components are opened. This method should be called
// before Close(). Otherwise Open() will return error. If Open() method of one
// of components returns error all opened components will be closed. This
// method may be called many times and will return equal results. It's
// guaranteed that Open() method of all components will be called only once.
func (c *composite) Open() (err error) {
	select {
	case <-c.control.ctx.Done(): // closed
		if !c.control.isOpen() {
			// closed before open
			return ErrPrematurelyClosed
		}
		return c.control.openError.get()
	default:
	}
	if !c.control.tryOpen() {
		return c.control.openError.get()
	}

	c.handler(c.control)
	return c.control.openError.get()
}

// Close initialises shutdown for all Components. This method may be called
// many times and will return equal results. It's guaranteed that Close()
// method of all components will be called only once.
func (c *composite) Close() (err error) {
	select {
	case <-c.control.ctx.Done():
		// already closed
		return c.control.closeError.get()
	default:
		c.control.cancelFunc()
		c.control.closeWg.Wait()
	}
	return c.control.closeError.get()
}

// Wait blocks until all components are exited. If one of Wait() method of one
// of Components is exited before Close() all opened components will be closed.
// This method may be called many times and will return equal results. It's
// guaranteed that Wait() method of all components will be called only once.
func (c *composite) Wait() (err error) {
	<-c.control.ctx.Done()
	c.control.waitWg.Wait()
	return c.control.waitError.get()
}
