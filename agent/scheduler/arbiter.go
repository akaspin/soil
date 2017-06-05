package scheduler

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"github.com/mitchellh/hashstructure"
	"sync"
)

type managerCallback func(reason error, environment map[string]string, mark uint64)

type Arbiter struct {
	*supervisor.Control
	log *logx.Log

	mu      *sync.Mutex
	sources map[string]*Source
	managed map[string]*ManagedPod
}

func NewArbiter(ctx context.Context, log *logx.Log, sources ...agent.Source) (a *Arbiter) {
	a = &Arbiter{
		Control: supervisor.NewControl(ctx),
		log:     log.GetLog("manager"),
		mu:      &sync.Mutex{},
		sources: map[string]*Source{},
		managed: map[string]*ManagedPod{},
	}
	for _, s := range sources {
		a.sources[s.Name()] = &Source{
			source: s,
			cache: map[string]string{},
		}
	}
	return
}

func (a *Arbiter) Open() (err error) {
	for _, s := range a.sources {
		n := s.source.Name()
		s.source.Register(func(active bool, env map[string]string) {
			a.onCallback(n, active, env)
		})
	}
	err = a.Control.Open()
	return
}

func (a *Arbiter) Register(name string, pod *manifest.Pod, fn managerCallback) {
	if pod == nil {
		go a.removePod(name, fn)
		return
	}
	go a.addPod(name, pod, fn)
}


func (a *Arbiter) addPod(name string, pod *manifest.Pod, fn managerCallback)  {
	a.mu.Lock()
	a.managed[name] = &ManagedPod{
		Pod: pod,
		Fn: fn,
	}
	a.mu.Unlock()
	for _, source := range a.sources {
		source.source.SubmitPod(name, pod.Constraint)
		a.log.Debugf("%s is registered on %s with %v", name, source.source.Name(), pod.Constraint)
	}
}

func (a *Arbiter) removePod(name string, fn managerCallback) {
	a.mu.Lock()
	delete(a.managed, name)
	a.mu.Unlock()
	for _, a := range a.sources {
		a.source.RemovePod(name)
	}
	fn(nil, nil, 0)
	a.log.Debugf("remove %s", name)
}

func (a *Arbiter) onCallback(source string, active bool, env map[string]string) {
	a.log.Debugf("got callback from source %s (active:%t) %v", source, active, env)

	a.mu.Lock()
	defer a.mu.Unlock()

	a.sources[source].cache = env
	a.sources[source].active = active

	if !active {
		return
	}

	// get data
	namespaces := map[string]int{}
	marked := map[string]string{}
	all := map[string]string{}
	for _, s := range a.sources {
		if s.active {
			for _, ns := range s.source.Namespaces() {
				namespaces[ns] = namespaces[ns] + 1
			}
			for k, v := range s.cache {
				key := s.source.Name() + "." + k
				all[key] = v
				if s.source.Mark() {
					marked[key] = v
				}
			}
		}
	}

	mark, _ := hashstructure.Hash(marked, nil)
	for n, managed := range a.managed {
		if namespaces[managed.Pod.Namespace] > 0 {
			checkErr := managed.Pod.Constraint.Check(all)
			a.log.Debugf("notify %s %v %v", n, checkErr, all)
			managed.Fn(checkErr, all, mark)
		}
	}
}

type ManagedPod struct {
	Pod       *manifest.Pod
	Fn        func(reason error, environment map[string]string, mark uint64)
}

type Source struct {
	source agent.Source
	active bool
	cache map[string]string
}

func (s *Source) CanManage(namespace string) (res bool) {
	return
}