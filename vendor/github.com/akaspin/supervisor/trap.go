package supervisor

import "context"

// Trap can be used as watchdog in supervisor tree.
type Trap struct {
	ctx     context.Context
	cancel  context.CancelFunc
	lastErr compositeError
}

// NewTrap returns new Trap bounded to given Context
func NewTrap(ctx context.Context) (t *Trap) {
	t = &Trap{}
	t.ctx, t.cancel = context.WithCancel(ctx)
	return t
}

// Trap accepts given error and closes Trap if error is not nil
func (t *Trap) Trap(err error) {
	if err != nil {
		select {
		case <-t.ctx.Done():
			return
		default:
			t.lastErr.set(err)
			t.cancel()
		}
	}
}

// Open opens Trap and never returns any errors
func (*Trap) Open() (err error) {
	return nil
}

// Close closes trap
func (t *Trap) Close() (err error) {
	t.cancel()
	return
}

// Wait returns last accepted error
func (t *Trap) Wait() (err error) {
	<-t.ctx.Done()
	return t.lastErr.get()
}
