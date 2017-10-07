package resource

import (
	"fmt"
	"io"
	"github.com/akaspin/soil/manifest"
	"sync"
	"github.com/akaspin/soil/agent/bus"
)

const (
	workerTest = "test"
)

type Worker interface {
	io.Closer
	GetConfig() Config
}

func GetWorker(config Config, allocated Allocations) (w Worker, err error) {
	switch config.Type {
	case workerTest:
		//w = NewTestWorker(config, allocated)
	default:
		err = fmt.Errorf(`unknown worker %v`, config)
	}
	return
}

type Allocation struct {
	PodName string
	Request *manifest.Resource
	Values map[string]string
}

type Allocations map[string]*Allocation


type TestWorker struct {
	config Config
	consumer bus.MessageConsumer
	mu sync.Mutex
	allocated Allocations
}

func NewTestWorker(config Config, consumer bus.MessageConsumer, allocated Allocations) (w *TestWorker) {
	w = &TestWorker{
		config: config,
		allocated: allocated,
	}
	return
}

// Submit requests
func (w *TestWorker) Submit(podName string, requests []*manifest.Resource) {

}

// Close worker
func (w *TestWorker) Close() error {
	return nil
}

func (w *TestWorker) GetConfig() Config {
	return w.config
}



