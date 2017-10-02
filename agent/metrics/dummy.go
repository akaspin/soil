package metrics

import (
	"fmt"
	"sort"
	"sync"
)

// Dummy reporter for testing purposes
type Dummy struct {
	name string
	tags []string

	mu   sync.Mutex
	Data map[string]interface{}
}

func NewDummy(name string, tags ...string) (r *Dummy) {
	r = &Dummy{
		name: name,
		tags: tags,
		Data: map[string]interface{}{},
	}
	return
}

func (r *Dummy) Count(name string, value int64, tags ...string) {
	sort.Strings(tags)
	line := fmt.Sprintf("count:%s:%v", name, tags)
	r.mu.Lock()
	defer r.mu.Unlock()
	var old int64
	if val, ok := r.Data[line]; ok {
		old = val.(int64)
	}
	r.Data[line] = old + value
}
