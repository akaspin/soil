package scheduler

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/registry"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
)

type Sink struct {
	*supervisor.Control
	log *logx.Log

	stateHolder allocation.StateHolder
	managedEvaluators []registry.ManagedEvaluator

	state     *SinkState
}

func NewSink(ctx context.Context, log *logx.Log, stateHolder allocation.StateHolder, managedEvaluators ...registry.ManagedEvaluator) (r *Sink) {
	r = &Sink{
		Control:   supervisor.NewControl(ctx),
		log:       log.GetLog("scheduler", "sink", "pods"),
		stateHolder: stateHolder,
		managedEvaluators: managedEvaluators,
	}
	return
}

func (s *Sink) Open() (err error) {
	dirty := map[string]string{}
	for _, recovered := range s.stateHolder.GetState() {
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
		s.submitToEvaluators(name, pod)
	}
	s.log.Infof("done: %s", namespace)
	return
}

func (s *Sink) submitToEvaluators(name string, pod *manifest.Pod) (err error) {
	for _, me := range s.managedEvaluators {
		if pod == nil {
			me.Manager.DeregisterResource(name, func() {
				me.Evaluator.Deallocate(name)
			})
			return
		}
		me.Manager.RegisterResource(name, pod.Namespace, me.Evaluator.GetConstraint(pod), func(reason error, env map[string]string, mark uint64) {
			s.log.Debugf("received %v from manager for %s", reason, name)
			if reason == nil {
				me.Evaluator.Allocate(name, pod, env)
			} else {
				me.Evaluator.Deallocate(name)
			}
		})
	}
	return
}
