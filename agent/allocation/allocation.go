package allocation

import (
	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/copystructure"
	"io"
	"strings"
)

// Slice of recoverable entities
type RecoverableSlice interface {
	Append(v Recoverable)
}

// Recoverable entity can recover own state
type Recoverable interface {
	RestoreState(line string) (err error)
	StoreState(w io.Writer) (err error)
}

func Recover(v RecoverableSlice, empty Recoverable, source string, prefixes []string) (err error) {
	err = &multierror.Error{}
	for _, line := range strings.Split(source, "\n") {
		for _, prefix := range prefixes {
			if strings.HasPrefix(line, prefix) {
				cp, _ := copystructure.Copy(empty)
				v1 := cp.(Recoverable)
				if rErr := v1.RestoreState(line); rErr != nil {
					err = multierror.Append(err, rErr)
					continue
				}
				v.Append(v1)
			}
		}
	}
	err = err.(*multierror.Error).ErrorOrNil()
	return
}
