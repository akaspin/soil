package bus

import (
	"github.com/mitchellh/hashstructure"
)

// Message
type Message struct {
	prefix  string            // Producer name
	payload map[string]string // Message payload
	mark    uint64
}

// Create new message non-expiring with cloned payload. Use <nil> payload to create empty message.
func NewMessage(prefix string, payload map[string]string) (m Message) {
	m = Message{
		prefix: prefix,
	}
	if payload != nil {
		m.payload = CloneMap(payload)
	}
	m.mark, _ = hashstructure.Hash(m.payload, nil)
	return
}

// Get message prefix
func (m Message) GetPrefix() string {
	return m.prefix
}

// Get message payload
func (m Message) GetPayload() map[string]string {
	if m.payload == nil {
		return map[string]string{}
	}
	return m.payload
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
	res = m.prefix == ingest.prefix && m.mark == ingest.mark
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
