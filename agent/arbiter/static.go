package arbiter

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"sync"
)

type StaticConfig struct {
	Id string
	Meta map[string]string
	PodExec string
	Constraint []*manifest.Pod
}

func (c StaticConfig) GetEnvironment() (res *agent.Environment) {
	fields := map[string]string{
		"agent.id": c.Id,
		"agent.pod.exec": c.PodExec,
	}
	for k, v := range c.Meta {
		fields["meta." + k] = v
	}
	res = agent.NewEnvironment(fields)
	return
}


// Static blocker ignores "count" parameter
type Static struct {
	*supervisor.Control
	log    *logx.Log
	config StaticConfig

	environment *agent.Environment

	state map[string]error
	mu *sync.Mutex
}


func NewStatic(ctx context.Context, log *logx.Log, config StaticConfig) (p *Static)  {
	p = &Static{
		Control: supervisor.NewControl(ctx),
		log: log.GetLog("blocker", "private"),
		config: config,
		environment: config.GetEnvironment(),
		state: map[string]error{},
		mu: &sync.Mutex{},
	}
	return
}

func (s *Static) Open() (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range s.config.Constraint {
		s.state[p.Name] = s.environment.Assert(p.Constraint)
	}
	s.log.Debugf("open %v", s.state)
	err = s.Control.Open()
	return
}

func (s *Static) Submit(name string, pod *manifest.Pod, fn func(reason error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if pod == nil {
		return
	}
	go fn(s.state[name])
	return
}

func (s *Static) Environment() (res *agent.Environment) {
	res = s.environment
	return
}

