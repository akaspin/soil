package source

import (
	"context"
	"github.com/akaspin/logx"
)

type baseSource struct {
	*BaseProducer
	namespaces []string
	mark       bool
}

func newBaseSource(ctx context.Context, log *logx.Log, prefix string, namespaces []string, mark bool) (s *baseSource) {
	s = &baseSource{
		BaseProducer: NewBaseProducer(ctx, log, prefix),
		namespaces:   namespaces,
		mark:         mark,
	}
	return
}

func (s *baseSource) Namespaces() []string {
	return s.namespaces
}

func (s *baseSource) Mark() bool {
	return s.mark
}
