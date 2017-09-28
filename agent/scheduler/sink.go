package scheduler

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
)

type Sink struct {
	*supervisor.Control
	log *logx.Log

	evaluator *Evaluator
	manager   *Manager
	state     *SinkState
}

func NewSink(ctx context.Context, log *logx.Log, evaluator *Evaluator, manager *Manager) (r *Sink) {
	r = &Sink{
		Control:   supervisor.NewControl(ctx),
		log:       log.GetLog("scheduler", "sink", "pods"),
		evaluator: evaluator,
		manager:   manager,
	}
	return
}

func (s *Sink) Open() (err error) {
	dirty := map[string]string{}
	for _, recovered := range s.evaluator.List() {
		dirty[recovered.Name] = recovered.Namespace
	}
	s.state = NewSinkState([]string{"private", "public"}, dirty)
	err = s.Control.Open()
	s.log.Debugf("open")
	return
}

// SyncNamespace scheduler pods. Called by registry on initialization.
func (s *Sink) ConsumeRegistry(namespace string, pods manifest.Registry) {
	s.log.Debugf("begin: %s", namespace)

	changes := s.state.SyncNamespace(namespace, pods)
	for name, pod := range changes {
		s.submitToEvaluator(name, pod)
	}
	s.log.Infof("done: %s", namespace)
	return
}

func (s *Sink) submitToEvaluator(name string, pod *manifest.Pod) (err error) {
	if pod == nil {
		go s.manager.DeregisterResource(name, func() {
			s.evaluator.Submit(name, nil)
		})
		return
	}
	s.manager.RegisterResource(name, pod.Namespace, pod.GetConstraint(), func(reason error, env map[string]string, mark uint64) {
		s.log.Debugf("received %v from manager for %s", reason, name)
		var alloc *allocation.Pod
		if pod != nil && reason == nil {
			if alloc, err = allocation.NewFromManifest(pod, allocation.DefaultSystemDPaths(), env); err != nil {
				return
			}
		}
		s.evaluator.Submit(name, alloc)
	})
	return
}
