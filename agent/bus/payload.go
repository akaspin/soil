package bus

import (
	"encoding/json"
	"github.com/mitchellh/copystructure"
	"github.com/mitchellh/hashstructure"
	"github.com/mitchellh/mapstructure"
)

// Immutable payload
type Payload interface {
	IsEmpty() bool
	Hash() uint64
	Unmarshal(interface{}) error
	JSON() ([]byte, error)
	Clone() Payload
}

func NewPayload(v interface{}) (p Payload) {
	switch v1 := v.(type) {
	case nil:
		p = NewFlatMapPayload(nil)
	case Payload:
		p = v1.Clone()
	case map[string]string:
		p = NewFlatMapPayload(v1)
	default:
		// interface
		data, _ := json.Marshal(v1)
		p = NewJSONPayload(data)
	}
	return
}

// Flat map payload
type FlatMapPayload struct {
	data map[string]string
	mark uint64
}

func NewFlatMapPayload(v map[string]string) (p FlatMapPayload) {
	if v == nil {
		return
	}
	v1, _ := copystructure.Copy(v)
	p.data = v1.(map[string]string)
	p.mark, _ = hashstructure.Hash(p.data, nil)
	return
}

func (p FlatMapPayload) IsEmpty() bool {
	return p.data == nil
}

func (p FlatMapPayload) Hash() uint64 {
	return p.mark
}

func (p FlatMapPayload) Unmarshal(v interface{}) (err error) {
	err = mapstructure.Decode(p.data, v)
	return
}

func (p FlatMapPayload) JSON() (res []byte, err error) {
	res, err = json.Marshal(&p.data)
	return
}

func (p FlatMapPayload) Clone() (res Payload) {
	res = NewFlatMapPayload(p.data)
	return
}

// JSON payload holds data in JSON
type JSONPayload struct {
	data []byte
	mark uint64
}

func NewJSONPayload(v []byte) (p JSONPayload) {
	if v == nil {
		return
	}
	p.data = make([]byte, len(v))
	copy(p.data, v)
	p.mark, _ = hashstructure.Hash(p.data, nil)
	return
}

func (p JSONPayload) IsEmpty() bool {
	return p.data == nil
}

func (p JSONPayload) Hash() uint64 {
	return p.mark
}

func (p JSONPayload) Unmarshal(v interface{}) error {
	return json.Unmarshal(p.data, v)
}

func (p JSONPayload) JSON() ([]byte, error) {
	return p.data, nil
}

func (p JSONPayload) Clone() Payload {
	return NewJSONPayload(p.data)
}
//
//func (p JSONPayload) String() (res string) {
//	res = string(p.data)
//	return
//}
