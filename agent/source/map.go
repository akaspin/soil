package source

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"sync"
)

// MapSource arbiter dynamically evaluates map of parameters
type MapSource struct {
	*supervisor.Control
	log        *logx.Log
	name       string
	namespaces []string
	mark       bool
	required   manifest.Constraint

	callback func(bool, map[string]string)
	fields   map[string]string
	active   bool
	mu       *sync.Mutex
}

func NewMapSource(ctx context.Context, log *logx.Log, name string, mark bool, required manifest.Constraint) (s *MapSource) {
	s = &MapSource{
		Control:    supervisor.NewControl(ctx),
		log:        log.GetLog("metadata", "map", name),
		name:       name,
		namespaces: []string{"private", "public"},
		mark:       mark,
		required:   required,
		callback:   func(bool, map[string]string) {},
		fields:     map[string]string{},
		mu:         &sync.Mutex{},
	}
	return
}

func (s *MapSource) Name() string {
	return s.name
}

func (s *MapSource) Namespaces() []string {
	return s.namespaces
}

func (s *MapSource) Mark() bool {
	return s.mark
}

func (s *MapSource) Required() manifest.Constraint {
	return s.required
}

func (s *MapSource) Register(callback func(active bool, env map[string]string)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callback = callback
	s.log.Debugf("register")
	return
}

func (s *MapSource) SubmitPod(name string, constraints manifest.Constraint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callback(s.active, s.fields)
}

func (s *MapSource) RemovePod(name string) {
}

func (s *MapSource) Configure(v map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = true
	s.fields = v
	s.log.Debugf("sync %v", v)
	s.callback(s.active, s.fields)
}

func (s *MapSource) Set(v map[string]string, replace bool) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = true
	if replace {
		s.fields = v
		s.log.Infof("replace %v", v)
	} else {
		for k, v1 := range v {
			s.fields[k] = v1
		}
		s.log.Infof("merge %v : %v", v, s.fields)
	}
	s.callback(s.active, s.fields)
	return
}

func (s *MapSource) Delete(keys ...string) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = true
	for _, k := range keys {
		delete(s.fields, k)
	}
	s.log.Infof("delete %v : %v", keys, s.fields)
	return
}

func (s *MapSource) Get() (v map[string]string, active bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v = s.fields
	active = s.active
	return
}
