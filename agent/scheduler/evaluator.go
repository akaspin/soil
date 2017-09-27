package scheduler

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/supervisor"
	"github.com/coreos/go-systemd/dbus"
	"sync"
)

type Evaluator struct {
	*supervisor.Control
	log       *logx.Log
	//reporters []agent.EvaluationReporter

	state  *EvaluatorState
	nextCh chan *Evaluation
}

func NewEvaluator(ctx context.Context, log *logx.Log) (e *Evaluator) {
	e = &Evaluator{
		Control:   supervisor.NewControl(ctx),
		log:       log.GetLog("scheduler", "evaluator"),
		//reporters: reporters,
		nextCh:    make(chan *Evaluation),
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
		e.log.Infof("restored allocation %v", alloc.Header)
	}
	e.state = NewEvaluatorState(res)
	//for _, reporter := range e.reporters {
	//	reporter.Sync(res)
	//}
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
	var failures []error
	e.log.Tracef("begin %s", evaluation)
	conn, err := dbus.New()
	if err != nil {
		e.log.Error(err)
		return
	}
	defer conn.Close()

	currentPhase := -1
	var phase []Instruction
	for _, instruction := range evaluation.Plan() {
		if currentPhase < instruction.Phase() {
			currentPhase = instruction.Phase()
			failures = append(failures, e.executePhase(phase, conn)...)
			phase = []Instruction{}
		}
		phase = append(phase, instruction)
	}
	failures = append(failures, e.executePhase(phase, conn)...)

	e.log.Infof("evaluation done %s (failures:%v)", evaluation, failures)
	next := e.state.Commit(evaluation.Name())
	//for _, reporter := range e.reporters {
	//	reporter.Report(evaluation.Name(), evaluation.Right, failures)
	//}
	e.fanOut(next)
	return
}

func (e *Evaluator) executePhase(phase []Instruction, conn *dbus.Conn) (failures []error) {
	if len(phase) == 0 {
		return
	}
	e.log.Debugf("begin phase %v", phase)
	ch := make(chan error, len(phase))
	wg := &sync.WaitGroup{}
	wg.Add(len(phase))
	for _, instruction := range phase {
		go func(instruction Instruction) {
			defer wg.Done()
			e.log.Tracef("begin instruction %v", instruction)
			var iErr error
			if iErr = instruction.Execute(conn); iErr != nil {
				e.log.Errorf("error while execute instruction %v: %s", instruction, iErr)
			}
			e.log.Tracef("finish instruction %s", instruction)
			ch <- iErr
		}(instruction)
	}
	go func() {
		for res := range ch {
			if res != nil {
				failures = append(failures, res)
			}
		}
	}()
	wg.Wait()
	e.log.Debugf("finish phase %v", phase)

	return
}

func (e *Evaluator) fanOut(next []*Evaluation) {
	if len(next) > 0 {
		e.Control.Acquire()
		go func() {
			defer e.Control.Release()
			for _, evaluation := range next {
				e.nextCh <- evaluation
			}
		}()
	}
}
