package resource

import (
	"github.com/akaspin/soil/agent/allocation"
	"io"
)

// Estimator estimates resources and sends results to downstream consumer
// in form map[string]string there values are:
//
// 	   allocated = true|false
//	   <key> = "<value>"
//
// Each estimator estimates resources for only one provider
type Estimator interface {
	io.Closer

	Allocate(request *allocation.Resource) (err error)
	Deallocate(name string) (err error)
}
