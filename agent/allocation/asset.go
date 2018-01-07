package allocation

import (
	"io"
)

// Slice of recoverable entities
type AssetSlice interface {
	GetVersionPrefix(version string) (prefix string)
	GetEmpty(paths SystemPaths) (empty Asset)
	AppendItem(v Asset)
}

// Asset entity can recover own state
type Asset interface {
	UnmarshalSpec(line string, spec Spec, paths SystemPaths) (err error)
	MarshalSpec(w io.Writer) (err error)
}
