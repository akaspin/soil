package supervisor_test

import (
	"context"
	"fmt"
	"github.com/akaspin/supervisor"
)

// LayeredControl uses two controls to fully control component lifecycle.
// Use this pattern only if you need complex logic because all supervisor
// primitives guarantees that Open(), Close() and Wait() methods of supervised
// components will be called only once.
type LayeredControl struct {
	*supervisor.Control // managed externally
	doneControl         *supervisor.Control
}

func NewLayeredControl(ctx context.Context) (c *LayeredControl) {
	c = &LayeredControl{
		Control:     supervisor.NewControl(ctx),
		doneControl: supervisor.NewControl(context.Background()),
	}
	go func() {
		<-c.Ctx().Done()
		fmt.Println("shutting down")
		c.doneControl.Close()
	}()
	return c
}

func (c *LayeredControl) Open() (err error) {
	if c.Control.IsOpen() {
		return nil
	}
	fmt.Println("opening")
	c.Control.Open()
	c.doneControl.Open()
	return nil
}

func (c *LayeredControl) Close() (err error) {
	if c.Control.IsClosed() {
		return nil
	}
	fmt.Println("closing")
	return c.Control.Close()
}

func (c *LayeredControl) Wait() (err error) {
	if c.doneControl.IsClosed() {
		return nil
	}
	c.doneControl.Wait()
	fmt.Println("exited")
	return
}

func ExampleControl_layered() {
	c := NewLayeredControl(context.Background())
	c.Open()
	doneChan := make(chan struct{})
	go func() {
		c.Wait()
		close(doneChan)
	}()
	c.Close()
	<-doneChan
	c.Wait()

	// Output:
	// opening
	// closing
	// shutting down
	// exited
}
