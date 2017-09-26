package api_v1_test

import "sync"

type fixtureBackend struct {
	mu     *sync.Mutex
	states []map[string]string
}


func newFixtureBackend() (b *fixtureBackend) {
	b = &fixtureBackend{
		mu: &sync.Mutex{},
	}
	return
}

func (b *fixtureBackend) Set(data map[string]string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.states = append(b.states, data)
}
