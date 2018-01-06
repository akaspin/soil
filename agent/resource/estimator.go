package resource

import (
	"context"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/resource/estimator"
	"io"
)

// Estimator estimates resources and sends results to downstream consumer
// in form id:map[string]string there values are:
//
// 	   allocated = true|false
//	   <key> = "<value>"
//
// Downstream should be notified about destroyed resources by message with
// empty payload.
type Estimator interface {

	// Get estimator uuid
	Results() (uid string, ctx context.Context, ch chan *estimator.Result)

	// Create resource and notify downstream
	Create(id string, request *allocation.Resource) (err error)

	// Update resource and notify downstream
	Update(id string, request *allocation.Resource) (err error)

	// Destroy resource and notify downstream
	Destroy(name string) (err error)

	// Destroy all resources without notify downstream
	Shutdown()
	io.Closer
}

func GetEstimator(globalConfig estimator.GlobalConfig, config estimator.Config) (e Estimator, err error) {
	switch config.Provider.Kind {
	case "blackhole":
		e = estimator.NewBlackhole(globalConfig, config)
	case "range":
		e = estimator.NewRange(globalConfig, config)
	default:
		e = estimator.NewInvalid(globalConfig, config)
	}
	return
}
