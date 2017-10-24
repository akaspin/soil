package bus

import "sync"

type Storage struct {
	name string
	mu         sync.Mutex
	consumed   map[string]Message
	downstream []Consumer
}

func NewStorage(name string, downstream ...Consumer) (s *Storage) {
	s = &Storage{
		name: name,
		consumed: map[string]Message{},
		downstream: downstream,
	}
	return
}

func (s *Storage) ConsumeMessage(message Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handle(message)
}

func (s *Storage) handle(message Message) {
	prefix := message.GetPrefix()
	old, ok := s.consumed[prefix]
	if ok {
		if old.IsEqual(message) {
			return
		}
		if message.IsEmpty() {
			delete(s.consumed, prefix)
		} else {
			s.consumed[prefix] = message
		}
	} else {
		if message.IsEmpty() {
			return
		}
		s.consumed[prefix] = message
	}
	payload := map[string]string{}
	for p, m := range s.consumed {
		for k, v := range m.GetPayload() {
			payload[p + "." + k] = v
		}
	}
	msg := NewMessage(s.name, payload)
	for _, downstream := range s.downstream {
		downstream.ConsumeMessage(msg)
	}
	return
}


