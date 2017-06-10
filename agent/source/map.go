package source

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/manifest"
)

// Map arbiter dynamically evaluates map of parameters
type Map struct {
	*baseSource
	required manifest.Constraint
	fields   map[string]string
	active   bool
}

func NewMap(ctx context.Context, log *logx.Log, name string, mark bool, required manifest.Constraint) (s *Map) {
	s = &Map{
		baseSource: newBaseSource(ctx, log, name, []string{"private", "public"}, mark),
		required:   required,
		fields:     map[string]string{},
	}
	return
}

func (s *Map) Required() manifest.Constraint {
	return s.required
}

func (s *Map) Register(callback func(active bool, env map[string]string)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callback = callback
	s.log.Debugf("register")
	return
}

func (s *Map) SubmitPod(name string, constraints manifest.Constraint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callback(s.active, s.fields)
}

func (s *Map) RemovePod(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callback(s.active, s.fields)
}

func (s *Map) Configure(v map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = true
	s.fields = v
	s.log.Debugf("sync %v", v)
	s.callback(s.active, s.fields)
}

func (s *Map) Set(v map[string]string, replace bool) (err error) {
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

func (s *Map) Delete(keys ...string) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = true
	for _, k := range keys {
		delete(s.fields, k)
	}
	s.log.Infof("delete %v : %v", keys, s.fields)
	return
}
