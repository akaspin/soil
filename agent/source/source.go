package source

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/supervisor"
	"sync"
)

type baseSource struct {
	*supervisor.Control
	log        *logx.Log
	name       string
	namespaces []string
	mark       bool

	callback func(bool, map[string]string)
	active   bool
	mu       *sync.Mutex
}

func newBaseSource(ctx context.Context, log *logx.Log, name string, namespaces []string, mark bool) (s *baseSource) {
	s = &baseSource{
		Control:    supervisor.NewControl(ctx),
		log:        log.GetLog("source", name),
		name:       name,
		namespaces: namespaces,
		mark:       mark,
		callback:   func(bool, map[string]string) {},
		mu:         &sync.Mutex{},
	}
	return
}

func (s *baseSource) Name() string {
	return s.name
}

func (s *baseSource) Namespaces() []string {
	return s.namespaces
}

func (s *baseSource) Mark() bool {
	return s.mark
}
