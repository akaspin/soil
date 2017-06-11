package scheduler

import (
	"context"
	"github.com/akaspin/concurrency"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/supervisor"
	"github.com/coreos/go-systemd/dbus"
	"github.com/pkg/errors"
)

type Executor struct {
	*supervisor.Control
	log *logx.Log

	// bounded worker pool
	pool      *concurrency.WorkerPool
	reporters []agent.AllocationReporter

	// bounded state
	state *ExecutorState
}

func NewExecutor(ctx context.Context, log *logx.Log, pool *concurrency.WorkerPool, reporters ...agent.AllocationReporter) (r *Executor) {
	r = &Executor{
		Control:   supervisor.NewControl(ctx),
		log:       log.GetLog("executor"),
		pool:      pool,
		reporters: reporters,
	}
	return
}

func (r *Executor) Open() (err error) {
	r.log.Debug("open")
	var restored []*allocation.Pod
	if restored, err = r.restoreState(); err != nil {
		return
	}
	r.state = NewExecutorState(restored)
	for _, reporter := range r.reporters {
		reporter.Sync(restored)
	}
	err = r.Control.Open()
	return
}

func (r *Executor) Close() (err error) {
	r.log.Debug("close")
	err = r.Control.Close()
	return
}

func (r *Executor) Submit(name string, candidate *allocation.Pod) {
	if r.state.Submit(name, candidate) {
		defer r.deploy(name)
		return
	}
	r.log.Debugf("skip submit %s %s", name, allocation.ToString(candidate))
}

// ListActual latest allocations
func (r *Executor) List() (res map[string]*allocation.Header) {
	res = r.state.ListActual()
	return
}

// Try promote pending
func (r *Executor) deploy(name string) {
	log := r.log.GetLog(r.log.Prefix(), "deploy", name)

	ready, active, err := r.state.Promote(name)
	if err != nil {
		log.Debugf("skip promote : %s", err)
		return
	}
	log.Debugf("begin %s->%s", allocation.ToString(ready), allocation.ToString(active))
	go r.pool.Execute(r.Control.Ctx(), func() {
		defer r.deploy(name)
		failures := r.execute(log, ready, active)
		var commitErr error
		if _, commitErr = r.state.Commit(name, failures); commitErr != nil {
			log.Errorf("can't commit %s", err)
		}
		log.Infof("done %s->%s %v", allocation.ToString(ready), allocation.ToString(active), failures)
		for _, reporter := range r.reporters {
			reporter.Report(name, active, failures)
		}
	})
}

func (r *Executor) execute(log *logx.Log, ready, active *allocation.Pod) (failures []error) {

	plan := Plan(ready, active)
	log.Debugf("begin plan %v", plan)
	conn, err := dbus.New()
	if err != nil {
		r.log.Error(err)
		failures = append(failures, err)
		return
	}
	defer conn.Close()

	for _, instruction := range plan {
		log.Debugf("begin %v", instruction)
		if iErr := instruction.Execute(conn); iErr != nil {
			iErr = errors.Wrapf(iErr, "error while execute instruction %v: %s", instruction, iErr)
			log.Error(iErr)
			failures = append(failures, iErr)
			continue
		}
		log.Debugf("done %v", instruction)
	}
	log.Debugf("plan done %v (failures:%v)", plan, failures)
	return
}

func (r *Executor) restoreState() (res []*allocation.Pod, err error) {
	log := r.log.GetLog(r.log.Prefix(), "restoreState")
	log.Debug("begin")
	conn, err := dbus.New()
	if err != nil {
		return
	}
	defer conn.Close()

	files, err := conn.ListUnitFilesByPatterns([]string{}, []string{"pod-*.service"})
	if err != nil {
		return
	}
	for _, record := range files {
		log.Debugf("begin %s", record.Path)
		var alloc *allocation.Pod
		var allocErr error
		if alloc, allocErr = allocation.NewFromSystemD(record.Path); allocErr != nil {
			log.Warningf("can't restoreState allocation from %s", record.Path)
			continue
		}
		res = append(res, alloc)
		log.Debugf("done %v", alloc.Header)
	}
	r.log.Debug("done")
	return
}