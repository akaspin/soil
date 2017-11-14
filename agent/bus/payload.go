package bus

import (
	"encoding/json"
	"hash/crc64"
)

type Payload struct {
	data    []byte
	isEmpty bool
}

func NewPayload(v interface{}) (p Payload) {
	switch v1 := v.(type) {
	case Payload:
		p = Payload{
			isEmpty: v1.isEmpty,
			data: make([]byte, len(v1.data)),
		}
		copy(p.data, v1.data)
	default:
		p = Payload{
			isEmpty: v == nil,
		}
		p.data, _ = json.Marshal(v)
	}
	return
}

func (p Payload) IsEmpty() bool {
	return p.isEmpty
}

func (p Payload) Hash() uint64 {
	if p.isEmpty {
		return 0
	}
	return crc64.Checksum(p.data, crc64.MakeTable(crc64.ECMA))
}

func (p Payload) Unmarshal(v interface{}) error {
	return json.Unmarshal(p.data, v)
}

func (p Payload) String() (res string) {
	res = string(p.data)
	return
}
