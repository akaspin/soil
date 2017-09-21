package scheduler

import (
	"context"
	"errors"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/metadata"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"github.com/mitchellh/hashstructure"
	"sync"
)

var drainError = errors.New("agent in drain state")

type Manager struct {
	*supervisor.Control
	log *logx.Log

	mu      *sync.Mutex
	drain   bool
	sources map[string]*ManagerSource
	managed map[string]*managerResource

	dirtyNamespaces     map[string]struct{}
	interpolatableCache map[string]string
	interpolatableMark  uint64
	containableCache    map[string]string
}

func NewManager(ctx context.Context, log *logx.Log, sources ...*ManagerSource) (m *Manager) {
	m = &Manager{
		Control:             supervisor.NewControl(ctx),
		log:                 log.GetLog("arbiter"),
		mu:                  &sync.Mutex{},
		drain:               false,
		sources:             map[string]*ManagerSource{},
		managed:             map[string]*managerResource{},
		dirtyNamespaces:     map[string]struct{}{},
		interpolatableCache: map[string]string{},
		containableCache:    map[string]string{},
	}
	for _, source := range sources {
		m.sources[source.producer.Prefix()] = source
		for _, ns := range source.namespaces {
			m.dirtyNamespaces[ns] = struct{}{}
		}
	}
	return
}

func (m *Manager) Open() (err error) {
	for _, s := range m.sources {
		s.producer.RegisterConsumer("manager", m)
	}
	err = m.Control.Open()
	return
}

// Add or update new manageable resource
func (m *Manager) RegisterResource(name, namespace string, constraint manifest.Constraint, notifyFn func(reason error, environment map[string]string, mark uint64)) {
	go func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		resource := &managerResource{
			Namespace:  namespace,
			Constraint: constraint,
			Fn:         notifyFn,
		}
		m.managed[name] = resource
		m.log.Infof("resource registered: %s %v", name, constraint)
		m.notifyResource(name, resource)
	}()
}

func (m *Manager) DeregisterResource(name string, notifyFn func()) {
	go func() {
		m.mu.Lock()
		delete(m.managed, name)
		m.mu.Unlock()
		notifyFn()
		m.log.Debugf("removed %s", name)
	}()
}

// Drain modifies arbiter drain constraint
func (m *Manager) Drain(state bool) {
	m.log.Infof("received drain request %v", state)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.drain = state
	for n, managed := range m.managed {
		m.notifyResource(n, managed)
	}
}

func (m *Manager) DrainState() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.drain
}

// Sync takes data from one of sources and evaluates all cached data
func (m *Manager) Sync(message metadata.Message) {
	m.log.Debugf("got message %v", message)

	m.mu.Lock()
	defer m.mu.Unlock()

	// update data in cache
	m.sources[message.Prefix].message = message

	m.interpolatableCache = map[string]string{}
	m.containableCache = map[string]string{}
	m.dirtyNamespaces = map[string]struct{}{}
	for sourcePrefix, source := range m.sources {
		if source.message.Clean {
			// add fields if active
			for k, v := range source.message.Data {
				key := sourcePrefix + "." + k
				m.containableCache[key] = v
				if !source.constraintOnly {
					m.interpolatableCache[key] = v
				}
			}
			continue
		}
		for _, ns := range source.namespaces {
			m.dirtyNamespaces[ns] = struct{}{}
		}
	}
	m.interpolatableMark, _ = hashstructure.Hash(m.interpolatableCache, nil)

	if !message.Clean {
		return
	}
	for n, managed := range m.managed {
		m.notifyResource(n, managed)
	}
}

func (m *Manager) notifyResource(name string, resource *managerResource) {
	if m.drain {
		m.log.Debugf("notify %s about agent in drain state", name)
		resource.Fn(drainError, map[string]string{}, 0)
		return
	}
	if _, ok := m.dirtyNamespaces[resource.Namespace]; !ok {
		var checkErr error = resource.Constraint.Check(m.containableCache)
		resource.Fn(checkErr, m.interpolatableCache, m.interpolatableMark)
		m.log.Debugf("resource notified %s %v %v", name, checkErr, m.containableCache)
	}
}

type managerResource struct {
	Namespace  string
	Constraint manifest.Constraint
	Fn         func(reason error, environment map[string]string, mark uint64)
}

type ManagerSource struct {
	producer       metadata.Producer
	constraintOnly bool
	namespaces     []string
	message        metadata.Message
}

func NewManagerSource(producer metadata.Producer, constraintOnly bool, namespaces ...string) (s *ManagerSource) {
	s = &ManagerSource{
		producer:       producer,
		constraintOnly: constraintOnly,
		namespaces:     namespaces,
	}
	return
}
