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

type Manager struct {
	*supervisor.Control
	log *logx.Log

	arbiters []agent.Arbiter

	mu *sync.Mutex
	managed map[string]*ManagedPod
	cache map[string]map[string]string
	marked map[string]bool
}

func NewManager(ctx context.Context, log *logx.Log, arbiters ...agent.Arbiter) (m *Manager) {
	m = &Manager{
		Control: supervisor.NewControl(ctx),
		log: log.GetLog("manager"),
		arbiters: arbiters,
		mu: &sync.Mutex{},
		managed: map[string]*ManagedPod{},
		cache: map[string]map[string]string{},
		marked: map[string]bool{},
	}
	return
}

func (m *Manager) Open() (err error) {
	for _, a := range m.arbiters {
		n := a.Name()
		m.cache[n], m.marked[n] = a.RegisterManager(func(env map[string]string) {
			m.onCallback(n, env)
		})
	}
	err = m.Control.Open()
	return
}

func (m *Manager) Register(name string, pod *manifest.Pod, fn managerCallback) {
	if pod == nil {
		go m.deregister(name, fn)
		return
	}
	go m.register(name, pod, fn)
}


func (m *Manager) register(name string, pod *manifest.Pod, fn managerCallback)  {
	m.mu.Lock()
	m.managed[name] = &ManagedPod{
		Pod: pod,
		Fn: fn,
	}
	m.mu.Unlock()
	interest := pod.Constraint.ExtractFields()
	for _, arbiter := range m.arbiters {
		arbiterName := arbiter.Name()
		interests := interest[arbiterName]
		arbiter.RegisterPod(name, interests)
		m.log.Debugf("%s is registered on %s arbiter with %v", name, arbiterName, interests)
	}
}

func (m *Manager) deregister(name string, fn managerCallback) {
	m.mu.Lock()
	delete(m.managed, name)
	m.mu.Unlock()
	for _, a := range m.arbiters {
		a.DeregisterPod(name)
	}
	fn(nil, nil, 0)
	m.log.Debugf("deregister %s", name)
}

func (m *Manager) onCallback(arbiterName string, env map[string]string) {
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