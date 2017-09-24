package metadata

type Message struct {
	Prefix string            // Producer prefix
	Data   map[string]string // Message payload
}

func NewMessage(prefix string, payload map[string]string) (m Message) {
	m = Message{
		Prefix: prefix,
		Data:   payload,
	}
	return
}

func (m Message) GetPrefix() string {
	return m.Prefix
}

func (m Message) GetPayload() map[string]string {
	return m.Data
}
