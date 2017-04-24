package registry_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/registry"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

type dummyScheduler struct {
	pods map[string]*manifest.Pod
	mu *sync.Mutex
}

func newDummyScheduler() (s *dummyScheduler) {
	s = &dummyScheduler{
		pods: map[string]*manifest.Pod{},
		mu: &sync.Mutex{},
	}
	return
}

func (s *dummyScheduler) Sync(pods []*manifest.Pod) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pods = map[string]*manifest.Pod{}
	for _, p := range pods {
		s.pods[p.Name] = p
	}
	return
}

func (s *dummyScheduler) Submit(name string, pod *manifest.Pod) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if pod == nil {
		delete(s.pods, name)
		return
	}
	s.pods[name] = pod
	return
}

func TestNewPrivate(t *testing.T) {
	config := registry.PrivateConfig{
	}
	ctx := context.Background()
	log := logx.GetLog("test")
	schedulerRt := newDummyScheduler()

	privateRegistry := registry.NewPrivate(ctx, log, schedulerRt, config)
	assert.NoError(t, privateRegistry.Open())
	assert.NoError(t, privateRegistry.Close())
	assert.NoError(t, privateRegistry.Wait())

}


