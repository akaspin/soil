package allocation

import (
	"encoding/json"
	"io"
)

type HeaderMarshaller interface {
	MarshalHeader(w io.Writer, encoder *json.Encoder) (err error)
}

type HeaderUnmarshaller interface {
	UnmarshalHeader(line string) (err error)
}
