package metadata

import (
	"context"
	"github.com/akaspin/logx"
)

// Plain arbiter dynamically evaluates map of parameters
type Plain struct {
	*BaseProducer
}

func NewPlain(ctx context.Context, log *logx.Log, name string, constraintOnly bool) (s *Plain) {
	s = &Plain{
		BaseProducer: NewBaseProducer(ctx, log, name),
	}
	return
}

func (s *Plain) Configure(v map[string]string) {
	s.Store(true, v)
}
