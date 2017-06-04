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

	sources []agent.Source

	mu *sync.Mutex
	managed map[string]*ManagedPod
	cache map[string]map[string]string
	marked map[string]bool
}

func NewArbiter(ctx context.Context, log *logx.Log, arbiters ...agent.Source) (m *Arbiter) {
	m = &Arbiter{
		Control: supervisor.NewControl(ctx),
		log:     log.GetLog("manager"),
		sources: arbiters,
		mu:      &sync.Mutex{},
		managed: map[string]*ManagedPod{},
		cache:   map[string]map[string]string{},
		marked:  map[string]bool{},
	}
	return
}

func (m *Arbiter) Open() (err error) {
	for _, a1 := range m.sources {
		n := a1.Name()
		m.cache[n], m.marked[n] = a1.Register(func(env map[string]string) {
			m.onCallback(n, env)
		})
	}
	err = m.Control.Open()
	return
}

func (m *Arbiter) Register(name string, pod *manifest.Pod, fn managerCallback) {
	if pod == nil {
		go m.removePod(name, fn)
		return
	}
	go m.addPod(name, pod, fn)
}


func (m *Arbiter) addPod(name string, pod *manifest.Pod, fn managerCallback)  {
	m.mu.Lock()
	m.managed[name] = &ManagedPod{
		Pod: pod,
		Fn: fn,
	}
	m.mu.Unlock()
	for _, source := range m.sources {
		source.SubmitPod(name, pod.Constraint)
		m.log.Debugf("%s is registered on %s with %v", name, source.Name(), pod.Constraint)
	}
}

func (m *Arbiter) removePod(name string, fn managerCallback) {
	m.mu.Lock()
	delete(m.managed, name)
	m.mu.Unlock()
	for _, a := range m.sources {
		a.RemovePod(name)
	}
	fn(nil, nil, 0)
	m.log.Debugf("remove %s", name)
}

func (m *Arbiter) onCallback(arbiterName string, env map[string]string) {
	flat := map[string]string{}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cache[arbiterName] = env
	forMark := map[string]map[string]string{}
	for arbiterName, arbiter := range m.cache {
		if m.marked[arbiterName] {
			forMark[arbiterName] = arbiter
		}
		for k, v := range arbiter {
			flat[arbiterName+"."+k] = v
		}
	}
	mark, _ := hashstructure.Hash(forMark, nil)
	for n, managed := range m.managed {
		checkErr := managed.Pod.Constraint.Check(flat)
		m.log.Debugf("notify %s %v %v", n, checkErr, flat)
		managed.Fn(checkErr, flat, mark)
	}
}

type ManagedPod struct {
	Pod       *manifest.Pod
	Fn        func(reason error, environment map[string]string, mark uint64)
}