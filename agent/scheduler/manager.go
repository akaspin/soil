package scheduler

import (
	"context"
	"errors"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"sync"
)

var inactiveNamespaceError = errors.New("inactive namespace")

type Manager struct {
	*supervisor.Control
	log  *logx.Log
	name string

	mu      sync.RWMutex
	drain   bool
	sources map[string]ManagerSource
	records map[string]bus.Message
	managed map[string]*managedEntity

	dirtyNamespaces map[string]struct{}
	interpolatable  map[string]string
	containable     map[string]string
}

func NewManager(ctx context.Context, log *logx.Log, name string, sources ...ManagerSource) (m *Manager) {
	m = &Manager{
		Control:         supervisor.NewControl(ctx),
		log:             log.GetLog("scheduler", "manager", name),
		name:            name,
		drain:           false,
		sources:         map[string]ManagerSource{},
		records:         map[string]bus.Message{},
		managed:         map[string]*managedEntity{},
		dirtyNamespaces: map[string]struct{}{},
		interpolatable:  map[string]string{},
		containable:     map[string]string{},
	}
	for _, source := range sources {
		m.sources[source.producer] = source
		m.records[source.producer] = bus.NewMessage(source.producer, nil)
		for _, ns := range source.namespaces {
			m.dirtyNamespaces[ns] = struct{}{}
		}
	}
	return
}

// Add or update new manageable resource
func (m *Manager) RegisterResource(name, namespace string, constraint manifest.Constraint, notifyFn func(reason error, message bus.Message)) {
	m.mu.Lock()
	entity := &managedEntity{
		namespace:  namespace,
		constraint: constraint,
		notifyFn:   notifyFn,
		checkErr:   inactiveNamespaceError,
		mark:       ^uint64(0),
	}
	m.managed[name] = entity
	m.log.Infof(`registered: "%s" (namespace: %s, constraint: %v)`, name, namespace, constraint)
	m.notifyResource(name, entity)
	m.mu.Unlock()
}

// Unregister resource
func (m *Manager) UnregisterResource(name string, notifyFn func()) {
	m.mu.Lock()
	delete(m.managed, name)
	m.mu.Unlock()
	notifyFn()
	m.log.Infof(`unregistered: "%s"`, name)
}

// ConsumeMessage takes data from one of sources and evaluates all cached data
func (m *Manager) ConsumeMessage(message bus.Message) {
	m.mu.Lock()
	m.log.Tracef("got message %v (dirty: %v)", message, m.dirtyNamespaces)

	var ok bool

	old, ok := m.records[message.GetPrefix()]
	if ok && old.GetMark() == message.GetMark() && old.IsEmpty() == message.IsEmpty() {
		m.log.Tracef(`skipping update: equal ingest: %v(new) == %v(old)`, message, old)
		m.mu.Unlock()
		return
	}
	if _, ok = m.sources[message.GetPrefix()]; !ok {
		m.log.Warningf(`skipping update: can not found source "%s"`)
		m.mu.Unlock()
		return
	}

	m.records[message.GetPrefix()] = message

	m.log.Tracef(`updating caches (dirty: %v)`, m.dirtyNamespaces)
	m.interpolatable = map[string]string{}
	m.containable = map[string]string{}
	m.dirtyNamespaces = map[string]struct{}{}
	for sourcePrefix, source := range m.sources {
		if source.required != nil || !m.records[sourcePrefix].IsEmpty() {
			for key, v := range m.records[sourcePrefix].Expand() {
				m.containable[key] = v
				if !source.constraintOnly {
					m.interpolatable[key] = v
				}
			}
			continue
		}
		for _, ns := range source.namespaces {
			m.dirtyNamespaces[ns] = struct{}{}
		}
	}
	m.log.Tracef("caches updated (dirty: %v): %v", m.dirtyNamespaces, m.containable)
	for n, managed := range m.managed {
		m.notifyResource(n, managed)
	}
	m.mu.Unlock()
}

func (m *Manager) notifyResource(name string, entity *managedEntity) {
	m.log.Tracef(`evaluating "%s" with constraint %v against %v`, name, entity.constraint, m.containable)

	var checkErr error
	for _, source := range m.sources {
		if source.required != nil {
			if checkErr = source.required.Check(m.containable); checkErr != nil {
				entity.notifyFn(checkErr, bus.NewMessage(m.name, nil))
				m.log.Warningf(`"%s" failed required check: %v`, name, checkErr)
				return
			}
		}
	}

	if _, ok := m.dirtyNamespaces[entity.namespace]; !ok {
		if checkErr = entity.constraint.Check(m.containable); checkErr != nil {
			entity.notifyFn(checkErr, bus.NewMessage(m.name, nil))
			m.log.Debugf(`"%s" failed check: %v`, name, checkErr)
			return
		}
		entity.notifyFn(nil, bus.NewMessage(m.name, m.interpolatable))
		m.log.Debugf(`"%s" passed all constraint checks`, name)
	}
}

type managedEntity struct {
	namespace  string
	constraint manifest.Constraint
	notifyFn   func(reason error, message bus.Message)
	checkErr   error
	mark       uint64
}

type ManagerSource struct {
	producer       string
	constraintOnly bool                // use source only for constraints
	namespaces     []string            // namespaces to manage
	required       manifest.Constraint // required constraint
}

func NewManagerSource(producer string, constraintOnly bool, required manifest.Constraint, namespaces ...string) (s ManagerSource) {
	s = ManagerSource{
		producer:       producer,
		constraintOnly: constraintOnly,
		namespaces:     namespaces,
		required:       required,
	}
	return
}
