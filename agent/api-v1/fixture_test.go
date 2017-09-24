package api_v1_test

import "sync"

type fixtureBackend struct {
	mu *sync.Mutex
	states []map[string]string
	ttl []bool
}


func newFixtureBackend() (b *fixtureBackend) {
	b = &fixtureBackend{
		mu: &sync.Mutex{},
	}
	return
}

func (b *fixtureBackend) Set(data map[string]string, withTTL bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.ttl = append(b.ttl, withTTL)
	b.states = append(b.states, data)
}
