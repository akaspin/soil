package concurrency

import (
	"context"
	"github.com/akaspin/supervisor"
)

type jobRequest struct {
	ctx context.Context
	fn func()
}

type Worker struct {
	*supervisor.Control
	jobCh chan jobRequest
}

func newWorker(control *supervisor.Control) (w *Worker) {
	w = &Worker{
		Control: control,
		jobCh: make(chan jobRequest),
	}
	return
}

func (w *Worker) Open() (err error) {
	w.Control.Open()
	w.Acquire()
	go w.loop()
	return
}

func (w *Worker) Execute(ctx context.Context, fn func()) {
	select {
	case <-w.Control.Ctx().Done():
	case <-ctx.Done():
	case w.jobCh<- jobRequest{
			ctx: ctx,
			fn: fn,
		}:
	}
}

func (w *Worker) loop() {
	defer w.Release()
	LOOP:
	for {
		select {
		case <-w.Control.Ctx().Done():
			break LOOP
		case job := <-w.jobCh:
			w.run(job)
		}
	}
}

func (w *Worker) run(job jobRequest) {
	w.Acquire()
	defer w.Release()

	jobDoneCh := make(chan struct{})

	go func() {
		defer close(jobDoneCh)
		job.fn()
		select {
		case <-w.Control.Ctx().Done():
		case <-job.ctx.Done():
		case jobDoneCh<- struct {}{}:
		}
	}()

	select {
	case <-w.Control.Ctx().Done():
	case <-job.ctx.Done():
	case <-jobDoneCh:
	}
}

// WorkerPool uses pool of workers to execute tasks
type WorkerPool struct {
	control      *supervisor.Control
	resourcePool *ResourcePool
	config       Config
}

func NewWorkerPool(ctx context.Context, config Config) (p *WorkerPool) {
	p = &WorkerPool{
		control: supervisor.NewControl(ctx),
		config: config,
	}
	p.resourcePool = NewResourcePool(ctx, config, p.factory)
	return
}

func (p *WorkerPool) Open() (err error) {
	if err = p.resourcePool.Open(); err != nil {
		return
	}
	err = p.control.Open()
	return
}

func (p *WorkerPool) Close() (err error) {
	if err = p.resourcePool.Close(); err != nil {
		return
	}
	err = p.control.Close()
	return
}


func (p *WorkerPool) Wait() (err error) {
	if err = p.resourcePool.Wait(); err != nil {
		return
	}
	err = p.control.Wait()
	return
}

// Take worker from pool
func (p *WorkerPool) Get(ctx context.Context) (w *Worker, err error) {
	r, err := p.resourcePool.Get(ctx)
	if err != nil {
		return
	}
	w = r.(*Worker)
	return
}

func (p *WorkerPool) Put(w *Worker) (err error) {
	err = p.resourcePool.Put(w)
	return
}

func (p *WorkerPool) Execute(ctx context.Context, fn func()) (err error) {
	r, err := p.resourcePool.Get(ctx)
	if err != nil {
		return
	}
	w := r.(*Worker)
	w.Execute(ctx, func() {
		p.control.Acquire()
		defer p.control.Release()
		defer p.resourcePool.Put(w)
		fn()
	})
	return
}


func (p *WorkerPool) factory() (r Resource, err error) {
	w := newWorker(supervisor.NewControlTimeout(p.control.Ctx(), p.config.CloseTimeout))
	err = w.Open()
	if err != nil {
		return
	}
	r = w
	return
}
