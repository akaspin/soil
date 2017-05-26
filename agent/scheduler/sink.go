package scheduler

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"sync"
)


type Sink struct {
	*supervisor.Control
	log *logx.Log

	executor  *Executor
	manager *Manager
	state *SinkState

	//pods map[string]*manifest.Pod

	// namespaces in order

	mu *sync.Mutex
}

func NewSink(ctx context.Context, log *logx.Log, executor *Executor, manager *Manager) (r *Sink) {
	r = &Sink{
		Control:   supervisor.NewControl(ctx),
		log:       log.GetLog("scheduler"),
		executor:  executor,
		manager: manager,
		mu:        &sync.Mutex{},
	}
	return
}

func (s *Sink) Open() (err error) {
	s.log.Debugf("open")
	dirty := map[string]string{}
	for _, allocation := range s.executor.List() {
		dirty[allocation.Name] = allocation.Namespace
	}
	s.state = NewSinkState([]string{"private", "public"}, dirty)
	err = s.Control.Open()
	return
}

func (s *Sink) Close() error {
	s.log.Debug("close")
	return s.Control.Close()
}

func (s *Sink) Wait() (err error) {
	err = s.Control.Wait()
	return
}

// SyncNamespace scheduler pods. Called by registry on initialization.
func (s *Sink) Sync(namespace string, pods []*manifest.Pod) (err error) {
	s.log.Debugf("Sync %s", namespace)
	s.mu.Lock()
	defer s.mu.Unlock()

	changes := s.state.SyncNamespace(namespace, pods)
	for name, pod := range changes {
		s.submitToExecutor(name, pod)
	}
	s.log.Infof("sync %s done", namespace)
	return
}

func (s *Sink) submitToExecutor(name string, pod *manifest.Pod) (err error) {
	if pod == nil {
		go s.manager.Register(name, nil, func(res error, env map[string]string, mark uint64) {
			s.executor.Submit(name, nil)
		})
		return
	}
	s.manager.Register(name, pod, func(reason error, env map[string]string, mark uint64) {
		s.log.Debugf("received %v from manager for %s", reason, name)
		var alloc *Allocation
		if pod != nil && reason == nil {
			if alloc, err = NewAllocationFromManifest(pod, env, mark); err != nil {
				return
			}
		}
		s.executor.Submit(name, alloc)
	})
	return
}

