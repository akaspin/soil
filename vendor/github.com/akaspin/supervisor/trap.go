package supervisor

import (
	"sync"
	"context"
)

// Trap can be used to cancel specific context on error
type Trap struct {
	cancel context.CancelFunc
	err error
	errMu *sync.Mutex
}

// NewTrap returns Trap bounded to specific context cancel
func NewTrap(cancel context.CancelFunc) (t *Trap) {
	t = &Trap{
		cancel: cancel,
		errMu: &sync.Mutex{},
	}
	return
}

// Err returns first error
func (t *Trap) Err() (err error) {
	t.errMu.Lock()
	defer t.errMu.Unlock()
	err = t.err
	return
}

// Catch cancel bounded context on non-nil error
func (t *Trap) Catch(err error) {
	t.errMu.Lock()
	defer t.errMu.Unlock()
	if t.err == nil {
		t.err = err
	}
	t.cancel()
}
