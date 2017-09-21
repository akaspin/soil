package scheduler

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"sync"
	"github.com/akaspin/soil/agent/metadata"
)

type Sink struct {
	*supervisor.Control
	log *logx.Log

	evaluator *Evaluator
	arbiter   *metadata.Manager
	state     *SinkState

	mu *sync.Mutex
}

func NewSink(ctx context.Context, log *logx.Log, evaluator *Evaluator, manager *metadata.Manager) (r *Sink) {
	r = &Sink{
		Control:   supervisor.NewControl(ctx),
		log:       log.GetLog("scheduler"),
		evaluator: evaluator,
		arbiter:   manager,
		mu:        &sync.Mutex{},
	}
	return
}

func (s *Sink) Open() (err error) {
	s.log.Debugf("open")
	dirty := map[string]string{}
	for _, recovered := range s.evaluator.List() {
		dirty[recovered.Name] = recovered.Namespace
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
	s.log.Infof("sync %s finished", namespace)
	return
}

func (s *Sink) submitToExecutor(name string, pod *manifest.Pod) (err error) {
	if pod == nil {
		go s.arbiter.DeregisterResource(name, func() {
			s.evaluator.Submit(name, nil)
		})
		return
	}
	s.arbiter.RegisterResource(name, pod.Namespace, pod.Constraint, func(reason error, env map[string]string, mark uint64) {
		s.log.Debugf("received %v from arbiter for %s", reason, name)
		var alloc *allocation.Pod
		if pod != nil && reason == nil {
			if alloc, err = allocation.NewFromManifest(pod, env); err != nil {
				return
			}
		}
		s.evaluator.Submit(name, alloc)
	})
	return
}
