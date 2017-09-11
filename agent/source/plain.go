package source

import (
	"context"
	"github.com/akaspin/logx"
)

// Plain arbiter dynamically evaluates map of parameters
type Plain struct {
	*baseSource
}

func NewPlain(ctx context.Context, log *logx.Log, name string, mark bool) (s *Plain) {
	s = &Plain{
		baseSource: newBaseSource(ctx, log, name, []string{"private", "public"}, mark),
	}
	return
}

func (s *Plain) Configure(v map[string]string) {
	s.Store(true, v)
}
