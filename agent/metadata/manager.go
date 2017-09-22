package metadata

import (
	"context"
	"errors"
	"github.com/akaspin/logx"
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
		s.producer.RegisterConsumer("manager", m.Sync)
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

// Sync takes data from one of sources and evaluates all cached data
func (m *Manager) Sync(message Message) {
	m.log.Tracef("got message %v", message)

	m.mu.Lock()
	defer m.mu.Unlock()

	// update data in cache
	m.sources[message.Prefix].message = message

	m.interpolatableCache = map[string]string{}
	m.containableCache = map[string]string{}
	m.dirtyNamespaces = map[string]struct{}{}

	for sourcePrefix, source := range m.sources {
		if source.required != nil || source.message.Clean {
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

	if m.sources[message.Prefix].required == nil && !message.Clean {
		return
	}
	for n, managed := range m.managed {
		m.notifyResource(n, managed)
	}
}

func (m *Manager) notifyResource(name string, resource *managerResource) {
	m.log.Tracef("evaluating resource %s against %v", name, m.containableCache)

	var checkErr error
	for _, source := range m.sources {
		if source.required != nil {
			if checkErr = source.required.Check(m.containableCache); checkErr != nil {
				resource.Fn(checkErr, m.interpolatableCache, m.interpolatableMark)
				m.log.Warningf("(required) resource %s notified: %v %v", name, checkErr, m.containableCache)
				return
			}
		}
	}

	if _, ok := m.dirtyNamespaces[resource.Namespace]; checkErr == nil && !ok {
		checkErr = resource.Constraint.Check(m.containableCache)
		resource.Fn(checkErr, m.interpolatableCache, m.interpolatableMark)
		m.log.Debugf("resource %s notified: %v %v", name, checkErr, m.containableCache)
	}
}

type managerResource struct {
	Namespace  string
	Constraint manifest.Constraint
	Fn         func(reason error, environment map[string]string, mark uint64)
}

type ManagerSource struct {
	producer       Producer            // bounded producer
	constraintOnly bool                // use source only for constraints
	namespaces     []string            // namespaces to manage
	message        Message             // last message
	required       manifest.Constraint // required constraint
}

func NewManagerSource(producer Producer, constraintOnly bool, required manifest.Constraint, namespaces ...string) (s *ManagerSource) {
	s = &ManagerSource{
		producer:       producer,
		constraintOnly: constraintOnly,
		namespaces:     namespaces,
		message: Message{
			Prefix: producer.Prefix(),
		},
		required: required,
	}

	return
}
