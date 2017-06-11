package scheduler

import (
	"github.com/akaspin/supervisor"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/concurrency"
	"github.com/coreos/go-systemd/dbus"
	"fmt"
	"context"
	"github.com/akaspin/soil/agent"
)

type Evaluator struct {
	*supervisor.Control
	log *logx.Log
	pool *concurrency.WorkerPool
	reporters []agent.AllocationReporter

	state *EvaluatorState
	nextCh chan *Evaluation
}

func NewEvaluator(ctx context.Context, log *logx.Log, pool *concurrency.WorkerPool, reporters ...agent.AllocationReporter) (e *Evaluator) {
	e = &Evaluator{
		Control: supervisor.NewControl(ctx),
		log: log.GetLog("scheduler", "evaluator"),
		pool: pool,
		reporters: reporters,
		nextCh: make(chan *Evaluation),
	}
	return
}

func (e *Evaluator) Open() (err error) {
	conn, err := dbus.New()
	if err != nil {
		return
	}
	defer conn.Close()

	files, err := conn.ListUnitFilesByPatterns([]string{}, []string{"pod-*.service"})
	if err != nil {
		return
	}
	var res []*allocation.Pod
	for _, record := range files {
		e.log.Debugf("restoring allocation from %s", record.Path)
		var alloc *allocation.Pod
		var allocErr error
		if alloc, allocErr = allocation.NewFromSystemD(record.Path); allocErr != nil {
			e.log.Warningf("can't restore allocation from %s", record.Path)
			continue
		}
		res = append(res, alloc)
		e.log.Debugf("restored allocation %v", alloc.Header)
	}
	e.state = NewEvaluatorState(res)
	for _, reporter := range e.reporters {
		reporter.Sync(res)
	}
	err = e.Control.Open()
	go e.loop()
	return
}

func (e *Evaluator) Close() (err error) {
	e.log.Debug("closing")
	err = e.Control.Close()
	return
}

func (e *Evaluator) Submit(name string, pod *allocation.Pod) {
	select {
	case <-e.Control.Ctx().Done():
		e.log.Warningf("reject submit %s : evaluator is closed", name)
	default:
	}
	next := e.state.Submit(name, pod)
	e.fanOut(next)
}

func (e *Evaluator) List() (res map[string]*allocation.Header) {
	res = e.state.List()
	return
}

func (e *Evaluator) loop() {
	e.Acquire()
	defer e.Release()
	LOOP:
	for {
		select {
		case <-e.Control.Ctx().Done():
			break LOOP
		case next := <-e.nextCh:
			go e.execute(next)
		}
	}
	return
}

func (e *Evaluator) execute(evaluation *Evaluation) {
	e.log.Debugf("executing evaluation %v", evaluation)
	var failures []error

	plan := evaluation.Plan()
	e.log.Debugf("begin plan %v", plan)
	conn, err := dbus.New()
	if err != nil {
		e.log.Error(err)
		return
	}
	defer conn.Close()

	for _, instruction := range plan {
		e.log.Debugf("begin %v", instruction)
		if iErr := instruction.Execute(conn); iErr != nil {
			iErr = fmt.Errorf("error while execute instruction %v: %s", instruction, iErr)
			e.log.Error(iErr)
			failures = append(failures, iErr)
			continue
		}
		e.log.Debugf("finished %v", instruction)
	}
	e.log.Debugf("plan finished %v (failures:%v)", plan, failures)
	next := e.state.Commit(evaluation.Name())
	for _, reporter := range e.reporters {
		reporter.Report(evaluation.Name(), evaluation.Right, failures)
	}
	e.fanOut(next)
	return
}

func (e *Evaluator) fanOut(next []*Evaluation) {
	if len(next) > 0 {
		e.Control.Acquire()
		go func() {
			defer e.Control.Release()
			for _, evaluation := range next {
				e.nextCh<- evaluation
			}
		}()
	}
}