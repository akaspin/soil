package supervisor

import (
	"context"
	"errors"
	"time"
)

var (
	ErrTimeout = errors.New("timeout exceeded")
)

// Timeout supervises shutdown process of own descendant
type Timeout struct {
	ctx       context.Context
	cancel    context.CancelFunc
	timeout   time.Duration
	component Component

	openChan chan struct{} // open chan closed once
	openErr  compositeError

	closedChan chan struct{} // Close() method of descendant
	closeErr   compositeError

	doneCtx    context.Context
	doneCancel context.CancelFunc
	doneErr    compositeError
}

// NewTimeout creates new Timeout
func NewTimeout(ctx context.Context, timeout time.Duration, component Component) (t *Timeout) {
	t = &Timeout{
		timeout:    timeout,
		component:  component,
		openChan:   make(chan struct{}),
		closedChan: make(chan struct{}),
	}
	t.ctx, t.cancel = context.WithCancel(ctx)
	t.doneCtx, t.doneCancel = context.WithCancel(context.Background())
	return t
}

// Open opens supervised component and return error if any
func (t *Timeout) Open() (err error) {
	select {
	case <-t.openChan:
		return t.openErr.get()
	default:
	}
	go func() {
		defer close(t.openChan)
		if openErr := t.component.Open(); openErr != nil {
			t.openErr.set(openErr)
			return
		}
		// supervise close
		go func() {
			<-t.ctx.Done()
			select {
			case <-t.doneCtx.Done(): // already closed
			default:
				if closeErr := t.component.Close(); closeErr != nil {
					t.closeErr.set(closeErr)
				}
				go func() {
					select {
					case <-time.After(t.timeout):
						t.doneErr.set(ErrTimeout)
						t.doneCancel()
					case <-t.doneCtx.Done():
					}
				}()
			}
			close(t.closedChan)
		}()

		// supervise wait
		go func() {
			if doneErr := t.component.Wait(); doneErr != nil {
				select {
				case <-t.doneCtx.Done():
				default:
					t.doneErr.set(doneErr)
				}
			}
			t.doneCancel()
			t.cancel()
		}()
	}()
	<-t.openChan
	return t.openErr.get()
}

// Close closes supervised component and starts timer
func (t *Timeout) Close() (err error) {
	t.cancel()
	<-t.closedChan
	return t.closeErr.get()
}

// Wait blocks until Wait() of supervised component is exited
// or timeout is reached.
func (t *Timeout) Wait() (err error) {
	<-t.doneCtx.Done()
	return t.doneErr.get()
}
