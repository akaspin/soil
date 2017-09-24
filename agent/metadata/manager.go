package metadata

import (
	"context"
	"errors"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"github.com/mitchellh/hashstructure"
	"sync"
	"sort"
	"fmt"
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
		log:                 log.GetLog("metadata", "manager"),
		mu:                  &sync.Mutex{},
		drain:               false,
		sources:             map[string]*ManagerSource{},
		managed:             map[string]*managerResource{},
		dirtyNamespaces:     map[string]struct{}{},
		interpolatableCache: map[string]string{},
		containableCache:    map[string]string{},
	}
	for _, source := range sources {
		m.sources[source.message.Prefix] = source
		for _, ns := range source.namespaces {
			m.dirtyNamespaces[ns] = struct{}{}
		}
	}
	return
}

func (m *Manager) Open() (err error) {
	m.log.Info("open")
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
		m.log.Infof("register: %s %v", name, constraint)
		m.notifyResource(name, resource)
	}()
}

func (m *Manager) DeregisterResource(name string, notifyFn func()) {
	go func() {
		m.mu.Lock()
		delete(m.managed, name)
		m.mu.Unlock()
		notifyFn()
		m.log.Infof("deregister: %s", name)
	}()
}

// Sync takes data from one of sources and evaluates all cached data
func (m *Manager) Sync(message Message) {
	m.log.Tracef("got message %v", message)

	m.mu.Lock()
	defer m.mu.Unlock()

	previousDirty := m.getDirtyState(m.dirtyNamespaces)

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


	if currentDirty := m.getDirtyState(m.dirtyNamespaces); previousDirty != currentDirty {
		m.log.Infof("dirty namespaces changed: %s->%s", previousDirty, currentDirty)
	}

	if m.sources[message.Prefix].required == nil && !message.Clean {
		return
	}
	for n, managed := range m.managed {
		m.notifyResource(n, managed)
	}
}

func (m *Manager) getDirtyState(states map[string]struct{}) (res string) {
	var currentDirty []string
	for ns := range states {
		currentDirty = append(currentDirty, ns)
	}
	sort.Strings(currentDirty)
	res = fmt.Sprintf("%v", currentDirty)
	return
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
	constraintOnly bool                // use source only for constraints
	namespaces     []string            // namespaces to manage
	message        Message             // last message
	required       manifest.Constraint // required constraint
}

func NewManagerSource(producer string, constraintOnly bool, required manifest.Constraint, namespaces ...string) (s *ManagerSource) {
	s = &ManagerSource{
		constraintOnly: constraintOnly,
		namespaces:     namespaces,
		message: Message{
			Prefix: producer,
		},
		required: required,
	}

	return
}
