package bus

import "github.com/mitchellh/hashstructure"

// Message
type Message struct {
	producer string            // Producer name
	payload  map[string]string // Message payload
	mark uint64
}

// Create new message with cloned payload
func NewMessage(producerName string, payload map[string]string) (m Message) {
	m = Message{
		producer: producerName,
		payload:  map[string]string{},
	}
	if payload != nil {
		m.payload = CloneMap(payload)
	}
	m.mark, _ = hashstructure.Hash(m.payload, nil)
	return
}

// Create new Message without clone payload
func NewMessageUnsafe(producerName string, payload map[string]string) (m Message) {
	m = Message{
		producer: producerName,
		payload:  payload,
	}
	m.mark, _ = hashstructure.Hash(m.payload, nil)
	return
}

func (m Message) GetProducer() string {
	return m.producer
}

func (m Message) GetPayload() map[string]string {
	return m.payload
}

func (m Message) GetMark() (res uint64) {
	res = m.mark
	return
}

// Get clone of message
func (m Message) Clone() (res Message) {
	m = Message{
		producer: m.producer,
		payload:  CloneMap(m.payload),
		mark: m.mark,
	}
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
