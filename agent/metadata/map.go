package metadata

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"sync"
)

// MapMetadata arbiter dynamically evaluates map of parameters
type MapMetadata struct {
	*supervisor.Control
	log    *logx.Log
	name   string
	marked bool

	callback func(map[string]string)
	fields   map[string]string
	mu       *sync.Mutex
}

func NewMapMetadata(ctx context.Context, log *logx.Log, name string, marked bool) (a *MapMetadata) {
	a = &MapMetadata{
		Control:  supervisor.NewControl(ctx),
		log:      log.GetLog("metadata", "map", name),
		name:     name,
		marked:   marked,
		callback: func(map[string]string) {},
		fields:   map[string]string{},
		mu:       &sync.Mutex{},
	}
	return
}

func (a *MapMetadata) Name() string {
	return a.name
}

func (a *MapMetadata) Register(callback func(env map[string]string)) (current map[string]string, marked bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.callback = callback
	a.log.Debugf("register")
	current = a.fields
	marked = a.marked
	return
}

func (a *MapMetadata) SubmitPod(name string, constraints manifest.Constraint) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.callback != nil {
		a.callback(a.fields)
	}
}

func (a *MapMetadata) RemovePod(name string) {
}

func (a *MapMetadata) Configure(v map[string]string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.fields = v
	a.log.Debugf("sync %v", v)
	if a.callback != nil {
		a.callback(a.fields)
	}
}
