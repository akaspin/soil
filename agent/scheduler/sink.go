package scheduler

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
)

type RegistryConsumer interface {
	ConsumeRegistry(namespace string, payload manifest.Registry)
}

type Sink struct {
	*supervisor.Control
	log *logx.Log

	managedEvaluators []ManagedEvaluator

	state *SinkState
}

func NewSink(ctx context.Context, log *logx.Log, state allocation.State, managedEvaluators ...ManagedEvaluator) (s *Sink) {
	s = &Sink{
		Control:           supervisor.NewControl(ctx),
		log:               log.GetLog("scheduler", "sink"),
		managedEvaluators: managedEvaluators,
	}
	dirty := map[string]string{}
	for _, recovered := range state {
		dirty[recovered.Name] = recovered.Namespace
	}
	s.state = NewSinkState([]string{"private", "public"}, dirty)
	return
}

// SyncNamespace scheduler pods. Called by registry on initialization.
func (s *Sink) ConsumeRegistry(namespace string, pods manifest.Registry) {
	s.log.Debugf("begin: %s", namespace)

	changes := s.state.SyncNamespace(namespace, pods)
	var report []string
	for name, pod := range changes {
		s.submitToEvaluators(name, pod)
		if pod != nil {
			report = append(report, fmt.Sprintf(`%s(ns:%s,mark:%d)`, name, pod.Namespace, pod.Mark()))
			continue
		}
		report = append(report, fmt.Sprintf(`%s(nil)`, name))
	}
	s.log.Infof("submitted changes: %v", report)
	return
}

func (s *Sink) submitToEvaluators(name string, pod *manifest.Pod) (err error) {
	s.log.Debugf("submitting %s to %d managers", name, len(s.managedEvaluators))
	for _, me := range s.managedEvaluators {
		go func(me ManagedEvaluator, pod *manifest.Pod) {
			if pod == nil {
				s.log.Tracef(`unregister "%s" from manager "%s"`, name, me.manager.name)
				me.manager.UnregisterResource(name, func() {
					me.evaluator.Deallocate(name)
				})
				return
			}

			s.log.Tracef(`register "%s" in manager "%s"`, name, me.manager.name)
			constraint := me.evaluator.GetConstraint(pod)
			me.manager.RegisterResource(name, pod.Namespace, constraint, func(reason error, message bus.Message) {
				s.log.Tracef(`received %v from manager "%s" for "%s"`, reason, me.manager.name, name)
				if reason != nil {
					me.evaluator.Deallocate(name)
					return
				}
				me.evaluator.Allocate(name, pod, message.GetPayload())
			})
		}(me, pod)
	}
	return
}
