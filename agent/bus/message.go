package bus

import "fmt"

// Message encapsulates topic with given payload
type Message struct {
	topic   string // Topic
	payload Payload
}

// Create new message
func NewMessage(topic string, payload interface{}) (m Message) {
	return Message{
		topic:   topic,
		payload: NewPayload(payload),
	}
}

// Get topic
func (m Message) Topic() string {
	return m.topic
}

// Get message payload
func (m Message) Payload() (res Payload) {
	return m.payload
}

// Is message equal to given message
func (m Message) IsEqual(ingest Message) (res bool) {
	return m.topic == ingest.topic && m.Payload().Hash() == ingest.Payload().Hash()
}

func (m Message) String() (res string) {
	return fmt.Sprintf("%s:%s", m.topic, m.payload)
}
