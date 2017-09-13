package source

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
)

// Plain arbiter dynamically evaluates map of parameters
type Plain struct {
	*baseSource
	fields map[string]string
	active bool
}

func NewPlain(ctx context.Context, log *logx.Log, name string, mark bool) (s *Plain) {
	s = &Plain{
		baseSource: newBaseSource(ctx, log, name, []string{"private", "public"}, mark),
		fields:     map[string]string{},
	}
	return
}

func (s *Plain) RegisterConsumer(name string, consumer agent.SourceConsumer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callback = consumer.Sync
	s.log.Debugf("register")
	return
}

func (s *Plain) Notify() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callback(s.name, s.active, s.fields)
}

func (s *Plain) Configure(v map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = true
	s.fields = v
	s.log.Debugf("sync %v", v)
	s.callback(s.name, s.active, s.fields)
}

func (s *Plain) Set(v map[string]string, replace bool) (err error) {
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
	s.callback(s.name, s.active, s.fields)
	return
}

func (s *Plain) Delete(keys ...string) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = true
	for _, k := range keys {
		delete(s.fields, k)
	}
	s.log.Infof("delete %v : %v", keys, s.fields)
	return
}
