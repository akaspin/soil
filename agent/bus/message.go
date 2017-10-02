package bus

import "github.com/mitchellh/hashstructure"

// Message
type Message struct {
	producer string            // Producer name
	payload  map[string]string // Message payload
}

func NewMessage(producerName string, payload map[string]string) (m Message) {
	m = Message{
		producer: producerName,
		payload:  map[string]string{},
	}
	if payload != nil {
		m.payload = CloneMap(payload)
	}
	return
}

func (m Message) GetProducer() string {
	return m.producer
}

func (m Message) GetPayload() map[string]string {
	return m.payload
}

func (m Message) GetMark() (res uint64) {
	res, _ = hashstructure.Hash(m.payload, nil)
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
