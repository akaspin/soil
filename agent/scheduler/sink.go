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

type Sink struct {
	*supervisor.Control
	log               *logx.Log
	boundedEvaluators []BoundedEvaluator

	state *SinkState
}

func (s *Sink) ConsumeMessage(message bus.Message) (err error) {
	var pods manifest.PodSlice
	if err = message.Payload().Unmarshal(&pods); err != nil {
		return
	}
	s.ConsumeRegistry(pods)
	return
}

func NewSink(ctx context.Context, log *logx.Log, state allocation.PodSlice, boundedEvaluators ...BoundedEvaluator) (s *Sink) {
	s = &Sink{
		Control:           supervisor.NewControl(ctx),
		log:               log.GetLog("scheduler", "sink"),
		boundedEvaluators: boundedEvaluators,
	}
	dirty := map[string]string{}
	for _, recovered := range state {
		dirty[recovered.Name] = recovered.Namespace
	}
	s.state = NewSinkState([]string{"private", "public"}, dirty)
	return
}

func (s *Sink) ConsumeRegistry(registry manifest.PodSlice) {
	byNamespace := map[string]manifest.PodSlice{}
	for _, pod := range registry {
		byNamespace[pod.Namespace] = append(byNamespace[pod.Namespace], pod)
	}

	for ns, r := range byNamespace {
		s.log.Debugf("submitting: %s", ns)
		changes := s.state.SyncNamespace(ns, r)
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
	}

}

func (s *Sink) submitToEvaluators(id string, pod *manifest.Pod) {
	s.log.Debugf("submitting %s to %d arbiters", id, len(s.boundedEvaluators))
	for _, me := range s.boundedEvaluators {
		func(me BoundedEvaluator, pod *manifest.Pod) {
			if pod == nil {
				s.log.Tracef(`unregister "%s"`, id)
				me.binder.Unbind(id, func() {
					me.evaluator.Deallocate(id)
				})
				return
			}

			s.log.Tracef(`register "%s"`, id)
			constraint := me.evaluator.GetConstraint(pod)
			me.binder.Bind(id, constraint, func(reason error, message bus.Message) {
				s.log.Tracef(`received %v for "%s"`, reason, id)
				if reason != nil {
					me.evaluator.Deallocate(id)
					return
				}
				var env map[string]string
				if err := message.Payload().Unmarshal(&env); err != nil {
					s.log.Error(err)
				}
				me.evaluator.Allocate(pod, env)
			})
		}(me, pod)
	}
	return
}
