package executor

import (
	"context"
	"github.com/akaspin/concurrency"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/scheduler/allocation"
	"github.com/akaspin/supervisor"
	"github.com/coreos/go-systemd/dbus"
	"github.com/pkg/errors"
)

type Executor struct {
	*supervisor.Control
	log *logx.Log

	pool  *concurrency.WorkerPool
	state *State
}

func New(ctx context.Context, log *logx.Log, pool *concurrency.WorkerPool) (r *Executor) {
	r = &Executor{
		Control: supervisor.NewControl(ctx),
		log:     log.GetLog("executor"),
		pool:    pool,
	}
	return
}

func (r *Executor) Open() (err error) {
	r.log.Debug("open")
	var allocs []*allocation.Allocation
	if allocs, err = r.restore(); err != nil {
		return
	}
	r.state = NewState(allocs)
	err = r.Control.Open()
	return
}

func (r *Executor) Close() (err error) {
	r.log.Debug("close")
	err = r.Control.Close()
	return
}

func (r *Executor) Submit(name string, candidate *allocation.Allocation) {
	if r.state.Submit(name, candidate) {
		defer r.deploy(name)
		return
	}
	r.log.Debugf("skip submit %s %s", name, allocation.AllocationToString(candidate))
}

// ListActual latest allocations
func (r *Executor) List(namespace string) (res map[string]*allocation.AllocationHeader) {
	res = r.state.ListActual(namespace)
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
	log.Debugf("begin %s->%s", allocation.AllocationToString(ready), allocation.AllocationToString(active))
	go r.pool.Execute(r.Control.Ctx(), func() {
		defer r.deploy(name)
		failures := r.execute(log, ready, active)
		var commitErr error
		if _, commitErr = r.state.Commit(name, failures); commitErr != nil {
			log.Errorf("can't commit %s", err)
		}
		log.Debugf("done %s->%s", allocation.AllocationToString(ready), allocation.AllocationToString(active))
	})
}

func (r *Executor) execute(log *logx.Log, ready, active *allocation.Allocation) (failures []error) {

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

func (r *Executor) restore() (res []*allocation.Allocation, err error) {
	log := r.log.GetLog(r.log.Prefix(), "restore")
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
		var alloc *allocation.Allocation
		var allocErr error
		if alloc, allocErr = RestoreAllocation(record.Path); allocErr != nil {
			log.Warningf("can't restore allocation from %s", record.Path)
			continue
		}
		res = append(res, alloc)
		log.Debugf("done %v", alloc.AllocationHeader)
	}
	r.log.Debug("done")
	return
}

func RestoreAllocation(path string) (res *allocation.Allocation, err error) {
	res = &allocation.Allocation{
		AllocationFile: &allocation.AllocationFile{
			Path: path,
		},
		AllocationHeader: &allocation.AllocationHeader{},
	}
	if err = res.AllocationFile.Read(); err != nil {
		return
	}
	if res.Units, err = res.AllocationHeader.Unmarshal(res.AllocationFile.Source); err != nil {
		return
	}

	for _, u := range res.Units {
		if err = u.AllocationFile.Read(); err != nil {
			return
		}
	}
	return
}