package agent

import (
	"github.com/akaspin/soil/manifest"
	"net/http"
)

type Scheduler interface {
	// SyncNamespace internal state with given manifests
	Sync(namespace string, pods []*manifest.Pod) (err error)
}

type Source interface {

	// Name returns arbiter name
	Name() string

	// Source namespaces
	Namespaces() []string

	// Mark state
	Mark() bool

	// Required constraints
	Required() manifest.Constraint

	// Bind consumer. Source source will call callback on
	// change states.
	Register(callback func(active bool, env map[string]string))

	SubmitPod(name string, constraints manifest.Constraint)

	RemovePod(name string)

	Get() (v map[string]string, active bool)
}

type Configurable interface {
	Set(v map[string]string, replace bool) (err error)
	Delete(keys ...string) (err error)
}

// HTTP endpoint
type Endpoint interface {

	// handle request and return result
	Handle(resp http.ResponseWriter, req *http.Request) (res interface{}, err error)
}
