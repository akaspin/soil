package resource

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"sync"
)

// Resource evaluator
type Evaluator struct {
	*supervisor.Control
	log *logx.Log
	consumer bus.MessageConsumer

	mu sync.Mutex
	records map[string]*evaluatorRecord
}

func NewEvaluator(ctx context.Context, log *logx.Log, state allocation.State, consumer bus.MessageConsumer) (e *Evaluator) {
	e = &Evaluator{
		Control: supervisor.NewControl(ctx),
		log: log.GetLog("resource", "evaluator"),
		consumer:consumer,
		records: map[string]*evaluatorRecord{},
	}
	e.recoverState(state)
	return
}

func (e *Evaluator) Open() (err error) {
	e.consumer.ConsumeMessage(bus.NewMessage(
		manifest.ResourceValuesPrefix,
		map[string]string{},
	))
	err = e.Control.Open()
	return
}

func (e *Evaluator) Configure(configs ...Config) {
}

func (*Evaluator) GetConstraint(pod *manifest.Pod) manifest.Constraint {
	return pod.GetResourceRequestConstraint()
}

func (*Evaluator) Allocate(name string, pod *manifest.Pod, env map[string]string) {
	println("allocate", name, pod)
}

func (*Evaluator) Deallocate(name string) {
	println("deallocate", name)
}

func (e *Evaluator) recoverState(state allocation.State) {
}

type evaluatorRecord struct {
	config      *Config
	worker Worker
	allocations map[string]*allocation.Resource
}

func (r *evaluatorRecord) Close() error {
	if r.worker != nil {
		return r.worker.Close()
	}
	return nil
}

func newEvaluatorRecord(config *Config) (r *evaluatorRecord) {
	r = &evaluatorRecord{
		config: config,
		allocations: map[string]*allocation.Resource{},
	}
	return
}
