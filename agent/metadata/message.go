package metadata

type Message struct {
	Prefix string            // Producer prefix
	Clean  bool              // Producer state
	Data   map[string]string // Message payload
}

func NewCleanMessage(prefix string, payload map[string]string) (m Message) {
	m = Message{
		Prefix: prefix,
		Clean:  true,
		Data:   payload,
	}
	return
}

func NewDirtyMessage(prefix string) (m Message) {
	m = Message{
		Prefix: prefix,
	}
	return
}

func (m Message) GetPrefix() string {
	return m.Prefix
}

func (m Message) IsClean() bool {
	return m.Clean
}

func (m Message) GetPayload() map[string]string {
	return m.Data
}
