package bus

import "fmt"

// Message encapsulates topic with given payload
type Message struct {
	topic   string // Topic
	payload Payload
}

// Create new message
func NewMessage(topic string, payload interface{}) (m Message) {
	m = Message{
		topic:   topic,
		payload: NewPayload(payload),
	}
	return
}

// Get topic
func (m Message) Topic() string {
	return m.topic
}

// Get message payload
func (m Message) Payload() (res Payload) {
	res = m.payload
	return
}

// Is message equal to given message
func (m Message) IsEqual(ingest Message) (res bool) {
	res = m.topic == ingest.topic && m.Payload().Hash() == ingest.Payload().Hash()
	return
}

func (m Message) String() (res string) {
	res = fmt.Sprintf("%s:%s", m.topic, m.payload)
	return
}
