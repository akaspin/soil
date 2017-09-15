package scheduler

import (
	"context"
	"errors"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"github.com/mitchellh/hashstructure"
	"sync"
	"github.com/akaspin/soil/agent/metadata"
)

type managerCallback func(reason error, environment map[string]string, mark uint64)

type Manager struct {
	*supervisor.Control
	log *logx.Log

	mu        *sync.Mutex
	drain     bool
	producers map[string]*managerSource
	managed   map[string]*managerResource
}

func NewManager(ctx context.Context, log *logx.Log) (m *Manager) {
	m = &Manager{
		Control:   supervisor.NewControl(ctx),
		log:       log.GetLog("arbiter"),
		mu:        &sync.Mutex{},
		drain:     false,
		producers: map[string]*managerSource{},
		managed:   map[string]*managerResource{},
	}
	return
}

// AddProducer should be called before Open
func (m *Manager) AddProducer(producer metadata.Producer, constraintOnly bool, namespaces ...string)  {
	m.producers[producer.Prefix()] = &managerSource{
		producer: producer,
		constraintOnly: constraintOnly,
		namespaces: namespaces,
	}
}

func (m *Manager) Open() (err error) {
	for _, s := range m.producers {
		s.producer.RegisterConsumer("arbiter", m)
	}
	err = m.Control.Open()
	return
}

// Add or update new manageable resource
func (m *Manager) Register(name string, pod *manifest.Pod, fn managerCallback) {
	go m.addPod(name, pod.Namespace, pod.Constraint, fn)
}

func (m *Manager) Deregister(name string, fn managerCallback) {
	go m.removePod(name, fn)
}

func (m *Manager) addPod(name, namespace string, constraint manifest.Constraint, fn managerCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.managed[name] = &managerResource{
		Namespace:  namespace,
		Constraint: constraint,
		Fn:         fn,
	}
	m.evaluate()
}

func (m *Manager) removePod(name string, fn managerCallback) {
	m.mu.Lock()
	delete(m.managed, name)
	m.mu.Unlock()
	fn(nil, nil, 0)
	m.log.Debugf("removed %s", name)
}

// Drain modifies arbiter drain constraint
func (m *Manager) Drain(state bool) {
	m.log.Infof("received drain request %v", state)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.drain = state
	m.evaluate()
}

func (m *Manager) DrainState() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.drain
}

// Sync takes data from one of producers and evaluates all cached data
func (m *Manager) Sync(message metadata.Message) {
	m.log.Debugf("got message %v", message)

	m.mu.Lock()
	defer m.mu.Unlock()

	// update data in cache
	m.producers[message.Prefix].message = message

	// not clean - do nothing
	if !message.Clean {
		return
	}
	m.evaluate()
}

func (m *Manager) evaluate() {

	// inactive namespaces
	inactive := map[string]struct{}{}
	all := map[string]string{}
	interpolatable := map[string]string{}

	if m.drain {
		// disable all pods if agent in drain state
		drainErr := errors.New("agent in drain state")
		for n, managed := range m.managed {
			m.log.Debugf("notify %s about agent in drain state", n)
			managed.Fn(drainErr, interpolatable, 0)
		}
		return
	}

	for sourcePrefix, s := range m.producers {
		if s.message.Clean {
			// add fields if active
			for k, v := range s.message.Data {
				key := sourcePrefix + "." + k
				all[key] = v
				if !s.constraintOnly {
					interpolatable[key] = v
				}
			}
			continue
		}
		for _, ns := range s.namespaces {
			inactive[ns] = struct{}{}
		}
	}

	mark, _ := hashstructure.Hash(interpolatable, nil)
	for n, managed := range m.managed {
		if _, ok := inactive[managed.Namespace]; !ok {
			var checkErr error = managed.Constraint.Check(all)
			m.log.Debugf("notify %s %v %v", n, checkErr, all)
			managed.Fn(checkErr, interpolatable, mark)
		}
	}
}

type managerResource struct {
	Namespace  string
	Constraint manifest.Constraint
	Fn         func(reason error, environment map[string]string, mark uint64)
}

type managerSource struct {
	producer metadata.Producer
	constraintOnly bool
	namespaces []string
	message metadata.Message
}

