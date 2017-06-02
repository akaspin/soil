package arbiter

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/supervisor"
	"sync"
)

// MapArbiter arbiter dynamically evaluates map of parameters
type MapArbiter struct {
	*supervisor.Control
	log    *logx.Log
	name   string
	marked bool

	callback func(map[string]string)
	fields   map[string]string
	mu       *sync.Mutex
}

func NewMapArbiter(ctx context.Context, log *logx.Log, name string, marked bool) (a *MapArbiter) {
	a = &MapArbiter{
		Control:  supervisor.NewControl(ctx),
		log:      log.GetLog("arbiter", "map", name),
		name:     name,
		marked:   marked,
		callback: func(map[string]string) {},
		fields:   map[string]string{},
		mu:       &sync.Mutex{},
	}
	return
}

func (a *MapArbiter) Name() string {
	return a.name
}

func (a *MapArbiter) RegisterManager(callback func(env map[string]string)) (current map[string]string, marked bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.callback = callback
	a.log.Debugf("register manager")
	current = a.fields
	marked = a.marked
	return
}

func (a *MapArbiter) SubmitPod(name string, constraints map[string]string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.callback != nil {
		a.callback(a.fields)
	}
}

func (a *MapArbiter) RemovePod(name string) {
}

func (a *MapArbiter) Configure(v map[string]string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.fields = v
	a.log.Debugf("sync %v", v)
	if a.callback != nil {
		a.callback(a.fields)
	}
}
