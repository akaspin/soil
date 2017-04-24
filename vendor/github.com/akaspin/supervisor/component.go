package supervisor

import (
	"io"
)

type Component interface {
	io.Closer

	Open() (err error)

	// Wait blocks until component is closed or error occurs
	Wait() (err error)
}

