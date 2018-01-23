package supervisor

import (
	"errors"
	"github.com/akaspin/errslice"
	"sync"
)

var (
	// ErrPrematurelyClosed notifies that Close() method
	// was called before Open()
	ErrPrematurelyClosed = errors.New("prematurely closed")
)

type compositeError struct {
	sync.Mutex
	error
}

func (e *compositeError) set(err error) {
	e.Lock()
	defer e.Unlock()
	e.error = errslice.Append(e.error, err)
}

func (e *compositeError) get() (err error) {
	e.Lock()
	defer e.Unlock()
	return e.error
}
