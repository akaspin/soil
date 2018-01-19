package supervisor

import (
	"context"
	"sync/atomic"
)

/*
Chain supervises Components in order. All supervised components are open
in FIFO order and closed in LIFO order.

Chain collects and returns error from corresponding Component methods. If more
than one Components returns errors they will be wrapped in MultiError.
*/
type Chain struct {
	*composite
	components []Component
}

// NewChain creates new Chain. Provided context manages whole Chain. Close
// Context is equivalent to call Chain.Close().
func NewChain(ctx context.Context, components ...Component) (c *Chain) {
	c = &Chain{
		components: components,
	}
	c.composite = newComposite(ctx, func(control *compositeControl) {
		_, cancel := context.WithCancel(context.Background())
		c.buildLink(cancel, c.components)
	})
	return c
}

func (c *Chain) buildLink(ascendantCancel context.CancelFunc, tail []Component) {
	if len(tail) == 0 {
		// supervise last chunk
		go func() {
			<-c.control.ctx.Done()
			ascendantCancel()
		}()
		return
	}
	component := tail[0]

	if openErr := component.Open(); openErr != nil {
		c.control.openError.set(openErr)
		ascendantCancel()
		c.control.cancelFunc()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	var waitExited uint32

	// supervise close
	c.control.closeWg.Add(1)
	go func() {
		defer c.control.closeWg.Done()
		<-ctx.Done()
		if atomic.CompareAndSwapUint32(&waitExited, 0, 1) {
			if closeErr := component.Close(); closeErr != nil {
				c.control.closeError.set(closeErr)
			}
		}
	}()

	// supervise wait
	c.control.waitWg.Add(1)
	go func() {
		defer c.control.waitWg.Done()
		if waitErr := component.Wait(); waitErr != nil {
			c.control.waitError.set(waitErr)
		}
		atomic.CompareAndSwapUint32(&waitExited, 0, 1)
		select {
		case <-c.control.ctx.Done(): // normal shutdown
		default: // abnormal shutdown we need close context and wait for descendants
			c.control.cancelFunc()
			<-ctx.Done()
		}
		ascendantCancel()
		cancel()
	}()

	c.buildLink(cancel, tail[1:])
}
