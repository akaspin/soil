package bus

import (
	"encoding/json"
	"github.com/mitchellh/copystructure"
	"github.com/mitchellh/hashstructure"
)

// Message
type Message struct {
	id      string      // Producer id
	payload interface{} // Message payload
	mark    uint64
}

// Create new message non-expiring with cloned payload. Use <nil> payload to create empty message.
func NewMessage(id string, payload interface{}) (m Message) {
	m = Message{
		id: id,
	}
	if payload != nil {
		m.payload, _ = copystructure.Copy(payload)
	}
	m.mark, _ = hashstructure.Hash(m.payload, nil)
	return
}

// Get message ID
func (m Message) GetID() string {
	return m.id
}

// Get message payload as map[string]string. This method returns empty map if payload is <nil> or type is differs.
func (m Message) GetPayloadMap() (res map[string]string) {
	res, ok := m.payload.(map[string]string)
	if m.payload == nil || !ok {
		res = map[string]string{}
		return
	}
	return
}

func (m Message) GetPayloadJSON() (res []byte) {
	res, _ = json.Marshal(m.payload)
	return
}

// Get message payload hash
func (m Message) GetPayloadMark() (res uint64) {
	res = m.mark
	return
}

// Is message payload equal to <nil>
func (m Message) IsEmpty() (res bool) {
	res = m.payload == nil
	return
}

func (m Message) IsEqual(ingest Message) (res bool) {
	res = m.id == ingest.id && m.mark == ingest.mark
	return
}

// Clone payload
func CloneMap(payload map[string]string) (res map[string]string) {
	res = make(map[string]string, len(payload))
	for k, v := range payload {
		res[k] = v
	}
	return
}
