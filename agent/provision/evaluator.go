package provision

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"github.com/coreos/go-systemd/dbus"
	"sync"
)

type EvaluatorConfig struct {
	SystemPaths    allocation.SystemPaths
	Recovery       allocation.PodSlice // recovery state
	StatusConsumer bus.Consumer        // consumer for "evaluation.<pod>.*"
}

type Evaluator struct {
	*supervisor.Control
	log    *logx.Log
	config EvaluatorConfig

	state *EvaluatorState
}

func NewEvaluator(ctx context.Context, log *logx.Log, config EvaluatorConfig) (e *Evaluator) {
	e = &Evaluator{
		Control: supervisor.NewControl(ctx),
		log:     log.GetLog("provision", "evaluator"),
		config:  config,
	}
	e.state = NewEvaluatorState(e.log, config.Recovery)
	return
}

func (e *Evaluator) Open() (err error) {
	resetData := map[string]map[string]string{}
	for _, recovered := range e.config.Recovery {
		resetData[recovered.Name] = map[string]string{
			"present": "true",
			"state":   "dirty",
		}
	}
	e.config.StatusConsumer.ConsumeMessage(bus.NewMessage("", resetData))
	err = e.Control.Open()
	return
}

// Returns all base constraints including resources
func (e *Evaluator) GetConstraint(pod *manifest.Pod) (res manifest.Constraint) {
	res = pod.Constraint.Clone()
	if len(pod.Resources) > 0 {
		c1 := manifest.Constraint{}
		for _, r := range pod.Resources {
			c1[fmt.Sprintf(`${resource.%s.%s.allocated}`, pod.Name, r.Name)] = "true"
		}
		res = res.Merge(c1)
	}
	return
}

func (e *Evaluator) Allocate(pod *manifest.Pod, env map[string]string) {

	alloc := &allocation.Pod{
		UnitFile: allocation.UnitFile{
			SystemPaths: e.config.SystemPaths,
		},
	}
	if err := alloc.FromManifest(pod, env); err != nil {
		e.log.Error(err)
		return
	}
	e.submitAllocation(pod.Name, alloc)
}

func (e *Evaluator) Deallocate(name string) {
	e.submitAllocation(name, nil)
}

func (e *Evaluator) submitAllocation(name string, pod *allocation.Pod) {
	next := e.state.Submit(name, pod)
	e.fanOut(next)
}

func (e *Evaluator) fanOut(next []*Evaluation) {
	if len(next) > 0 {
		for _, evaluation := range next {
			go e.executeEvaluation(evaluation)
		}
	}
}

func (e *Evaluator) executeEvaluation(evaluation *Evaluation) {
	var failures []error
	e.log.Tracef("begin: %s", evaluation)
	conn, err := dbus.New()
	if err != nil {
		e.log.Error(err)
		return
	}
	defer conn.Close()

	plan := evaluation.Plan()
	name := evaluation.Name()
	var phase []Instruction
	currentPhase := -1

	state := "update"
	if evaluation.Right == nil {
		state = "destroy"
	} else if evaluation.Left == nil {
		state = "create"
	}
	e.config.StatusConsumer.ConsumeMessage(bus.NewMessage(name, map[string]string{
		"present": "true",
		"state":   state,
	}))

	for _, instruction := range plan {
		if currentPhase < instruction.Phase() {
			currentPhase = instruction.Phase()
			failures = append(failures, e.executePhase(phase, conn)...)
			phase = []Instruction{}
		}
		phase = append(phase, instruction)
	}
	failures = append(failures, e.executePhase(phase, conn)...)

	e.log.Debugf("plan done: %s:%s (failures:%v)", evaluation, plan, failures)
	e.log.Infof("evaluation done: %s (failures:%v)", evaluation, failures)
	if evaluation.Right != nil {
		e.config.StatusConsumer.ConsumeMessage(bus.NewMessage(name, map[string]string{
			"present": "true",
			"state":   "done",
		}))
	} else {
		e.config.StatusConsumer.ConsumeMessage(bus.NewMessage(evaluation.Name(), nil))
	}

	next := e.state.Commit(evaluation.Name())
	e.fanOut(next)
	return
}

func (e *Evaluator) executePhase(phase []Instruction, conn *dbus.Conn) (failures []error) {
	if len(phase) == 0 {
		return
	}
	e.log.Tracef("begin phase %v", phase)
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
