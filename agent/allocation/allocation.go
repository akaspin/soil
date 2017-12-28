package allocation

import (
	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/copystructure"
	"io"
	"strings"
)

// Slice of recoverable entities
type RecoverableSlice interface {
	Append(v ItemUnmarshaller)
}

// ItemUnmarshaller entity can recover own state
type ItemUnmarshaller interface {
	UnmarshalLine(line string) (err error)
}

// Marshal item to pod manifest
type ItemMarshaller interface {
	MarshalLine(w io.Writer) (err error)
}

// Recover items from pod unit header
func Recover(v RecoverableSlice, empty ItemUnmarshaller, source string, prefixes []string) (err error) {
	err = &multierror.Error{}
	for _, line := range strings.Split(source, "\n") {
		for _, prefix := range prefixes {
			if strings.HasPrefix(line, prefix) {
				cp, _ := copystructure.Copy(empty)
				v1 := cp.(ItemUnmarshaller)
				if rErr := v1.UnmarshalLine(line); rErr != nil {
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
