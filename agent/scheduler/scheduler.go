package scheduler

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/soil/agent/scheduler/allocation"
	"github.com/akaspin/soil/agent/scheduler/executor"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"sync"
)

type Runtime struct {
	*supervisor.Control
	log *logx.Log

	executorRt *executor.Runtime
	blocker agent.Filter
	namespace string

	pods map[string]*manifest.Pod
	mu *sync.Mutex
}

func (r *Runtime) Open() (err error) {
	r.log.Debugf("open")
	err = r.Control.Open()
	return
}

func (r *Runtime) Close() error {
	r.log.Debug("close")
	return r.Control.Close()
}

func (r *Runtime) Wait() (err error) {
	err = r.Control.Wait()
	return
}

func NewRuntime(ctx context.Context, log *logx.Log, executorRt *executor.Runtime, blocker agent.Filter, namespace string) (r *Runtime) {
	r = &Runtime{
		Control:    supervisor.NewControl(ctx),
		log:        log.GetLog("scheduler", namespace),
		executorRt: executorRt,
		blocker: blocker,
		namespace: namespace,
		pods: map[string]*manifest.Pod{},
		mu: &sync.Mutex{},
	}
	return
}

// Sync scheduler state. Called by registry on initialization.
func (r *Runtime) Sync(pods []*manifest.Pod) (err error) {
	r.log.Debug("Sync")
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range pods {
		r.pods[p.Name] = p
	}
	state := r.executorRt.List(r.namespace)

	// clean non-existent pods
	for k := range state {
		if _, ok := r.pods[k]; !ok {
			go r.executorRt.Submit(k, nil)
		}
	}
	for k, v := range r.pods {
		if allocErr := r.allocate(k, v); allocErr != nil {
			r.log.Errorf("can not allocate %s : %s", k, allocErr)
			continue
		}
	}
	r.log.Info("sync done")
	return
}

// Submit single pod
func (r *Runtime) Submit(name string, pod *manifest.Pod) (err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	err = r.allocate(name, pod)
	return
}

func (r *Runtime) allocate(name string, pod *manifest.Pod) (err error) {
	if pod != nil {
		r.pods[name] = pod
	} else {
		delete(r.pods, name)
	}
	r.blocker.Submit(name, pod, func(reason error) {
		r.log.Debugf(">>> %s %s", name, reason)
		var alloc *allocation.Pod
		if pod != nil && reason == nil {
			if alloc, err = allocation.NewFromManifest(r.namespace, pod, r.blocker.Environment()); err != nil {
				return
			}
		}
		r.executorRt.Submit(name, alloc)
	})

	return
}

