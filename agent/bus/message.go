package bus

// Message
type Message struct {
	id   string // Producer id
	data Payload
	ok   bool
}

// Create new message
func NewMessage(id string, payload interface{}) (m Message) {
	m = Message{
		id:   id,
		data: NewPayload(payload),
		ok:   true,
	}
	return
}

// Get message ID
func (m Message) GetID() string {
	return m.id
}

func (m Message) Payload() (res Payload) {
	if !m.ok {
		res = NewPayload(nil)
		return
	}
	res = m.data
	return
}

func (m Message) IsEqual(ingest Message) (res bool) {
	res = m.id == ingest.id && m.Payload().Hash() == ingest.Payload().Hash()
	return
}
