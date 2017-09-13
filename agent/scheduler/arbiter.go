package scheduler

import (
	"context"
	"errors"
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
	drain   bool
	sources map[string]*Source
	managed map[string]*ManagedPod
}

func NewArbiter(ctx context.Context, log *logx.Log, sources ...agent.Source) (a *Arbiter) {
	a = &Arbiter{
		Control: supervisor.NewControl(ctx),
		log:     log.GetLog("arbiter"),
		mu:      &sync.Mutex{},
		drain:   false,
		sources: map[string]*Source{},
		managed: map[string]*ManagedPod{},
	}
	for _, s := range sources {
		a.sources[s.Prefix()] = &Source{
			source: s,
			cache:  map[string]string{},
		}
	}
	return
}

func (a *Arbiter) Open() (err error) {
	for _, s := range a.sources {
		s.source.RegisterConsumer("arbiter", a)
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

func (a *Arbiter) addPod(name string, pod *manifest.Pod, fn managerCallback) {
	a.mu.Lock()
	a.managed[name] = &ManagedPod{
		Pod: pod,
		Fn:  fn,
	}
	a.evaluate()
	a.mu.Unlock()

	for _, source := range a.sources {
		source.source.Notify()
		a.log.Debugf("%s is registered on %s with %v", name, source.source.Prefix(), pod.Constraint)
	}
}

func (a *Arbiter) removePod(name string, fn managerCallback) {
	a.mu.Lock()
	delete(a.managed, name)
	a.mu.Unlock()
	for _, a := range a.sources {
		a.source.Notify()
	}
	fn(nil, nil, 0)
	a.log.Debugf("removed %s", name)
}

// Drain modifies arbiter drain constraint
func (a *Arbiter) Drain(state bool) {
	a.log.Infof("received drain request %v", state)
	a.mu.Lock()
	defer a.mu.Unlock()
	a.drain = state
	a.evaluate()
}

// Sync takes data from one of producers and evaluates all cached data
func (a *Arbiter) Sync(source string, active bool, data map[string]string) {
	a.log.Debugf("got callback from source %s (active:%t) %v", source, active, data)

	a.mu.Lock()
	defer a.mu.Unlock()

	// update data in cache
	a.sources[source].cache = data
	a.sources[source].active = active

	// not active - do nothing
	if !active {
		return
	}
	a.evaluate()
}

func (a *Arbiter) evaluate() {

	// inactive namespaces
	inactive := map[string]struct{}{}


	marked := map[string]string{}
	all := map[string]string{}

	if a.drain {
		// disable all pods if agent in drain state
		drainErr := errors.New("agent in drain state")
		for n, managed := range a.managed {
			a.log.Debugf("notify %s about agent in drain state", n)
			managed.Fn(drainErr, marked, 0)
		}
		return
	}

	for sourcePrefix, s := range a.sources {
		if s.active {
			// add fields if active
			for k, v := range s.cache {
				key := sourcePrefix + "." + k
				all[key] = v
				if s.source.Mark() {
					marked[key] = v
				}
			}
			continue
		}
		for _, ns := range s.source.Namespaces() {
			inactive[ns] = struct{}{}
		}
	}

	mark, _ := hashstructure.Hash(marked, nil)
	for n, managed := range a.managed {
		if _, ok := inactive[managed.Pod.Namespace]; !ok {
			var checkErr error = managed.Pod.Constraint.Check(all)
			a.log.Debugf("notify %s %v %v", n, checkErr, all)
			managed.Fn(checkErr, marked, mark)
		}
	}
}

type ManagedPod struct {
	Pod *manifest.Pod
	Fn  func(reason error, environment map[string]string, mark uint64)
}

type Source struct {
	source agent.Source
	active bool
	cache  map[string]string
}
