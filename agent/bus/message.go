package bus

// Message
type Message struct {
	id   string // Producer id
	data Payload
}

// Create new message
func NewMessage(id string, payload interface{}) (m Message) {
	m = Message{
		id:   id,
		data: NewPayload(payload),
	}
	return
}

// Get message ID
func (m Message) GetID() string {
	return m.id
}

func (m Message) Payload() (res Payload) {
	if m.data == nil {
		res = NewFlatMapPayload(nil)
		return
	}
	res = m.data
	return
}

func (m Message) IsEqual(ingest Message) (res bool) {
	res = m.id == ingest.id && m.data.Hash() == ingest.data.Hash()
	return
}
