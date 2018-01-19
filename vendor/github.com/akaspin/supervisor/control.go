package supervisor

import (
	"context"
)

// Control provides ability to turn any type to supervisor component.
//
//	type MyComponent struct {
//		*Control
//	}
//
//	myComponent := &MyComponent{
//		Control: NewControl(context.Background())
//	}
//
type Control struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// NewControl returns new Control
func NewControl(ctx context.Context, blockers ...Component) (c *Control) {
	c = &Control{}
	c.ctx, c.cancel = context.WithCancel(ctx)
	return
}

func (c *Control) Open() (err error) {
	return nil
}

// Close closes Control context
func (c *Control) Close() (err error) {
	c.cancel()
	return nil
}

// Wait blocks until component shutdown
func (c *Control) Wait() (err error) {
	<-c.ctx.Done()
	return nil
}

// Ctx returns Control context. Control context is always
// closed before blockers evaluation.
func (c *Control) Ctx() context.Context {
	return c.ctx
}
