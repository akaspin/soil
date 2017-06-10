package source

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/manifest"
	"strings"
)

// Status accepts reports from executor and provides
//
//     <pod> = present
//     <pod>.failures = [<failure>,failure..]
//     <pod>.namespace = private | public
//
type Status struct {
	*baseSource
	data map[string]string
}

func NewStatus(ctx context.Context, log *logx.Log) (s *Status) {
	s = &Status{
		baseSource: newBaseSource(ctx, log, "status", []string{"private", "public"}, false),
		data:       map[string]string{},
	}
	return
}

func (s *Status) Required() manifest.Constraint {
	return manifest.Constraint{}
}

func (s *Status) Register(callback func(active bool, env map[string]string)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callback = callback
	s.callback(s.active, s.data)
}

func (s *Status) SubmitPod(name string, constraints manifest.Constraint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callback(s.active, s.data)
}

func (s *Status) RemovePod(name string) {}

func (s *Status) Sync(pods []*allocation.Pod) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = map[string]string{}
	for _, v := range pods {
		s.update(v.Name, v, nil)
	}
	s.active = true
	s.callback(true, s.data)
}

func (s *Status) Report(name string, pod *allocation.Pod, failures []error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.update(name, pod, failures)
	s.callback(s.active, s.data)
}

func (s *Status) update(name string, pod *allocation.Pod, failures []error) {
	if pod == nil {
		delete(s.data, name)
		for k := range s.data {
			if strings.HasPrefix(k, name+".") {
				delete(s.data, k)
			}
		}
		return
	}
	s.data[name] = "present"
	s.data[name+".namespace"] = pod.Namespace
	s.data[name+".mark"] = fmt.Sprintf("%d", pod.PodMark)
	s.data[name+".agent_mark"] = fmt.Sprintf("%d", pod.AgentMark)
	s.data[name+".failures"] = fmt.Sprintf("%v", failures)
	var units []string
	for _, unit := range pod.Units {
		units = append(units, unit.UnitName())
	}
	s.data[name+".units"] = strings.Join(units, ",")
}
