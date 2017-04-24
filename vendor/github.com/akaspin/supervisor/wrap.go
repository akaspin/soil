package supervisor

import (
	"context"
	"sync"
)

type Wrapped struct {
	ctx context.Context
	cancel context.CancelFunc
	wg *sync.WaitGroup
	fn func() error
	trap *Trap
}

func NewWrapped(ctx context.Context, fn func() error) (w *Wrapped) {
	w = &Wrapped{
		wg: &sync.WaitGroup{},
		fn: fn,
	}
	w.ctx, w.cancel = context.WithCancel(ctx)
	w.trap = NewTrap(w.cancel)
	return
}

func (w *Wrapped) Open() (err error) {
	w.wg.Add(1)

	go func() {
		defer w.wg.Done()

		trackCh := make(chan error)
		go func() {
			defer close(trackCh)
			err := w.fn()
			trackCh <- err
		}()

		for {
			select {
			case err := <-trackCh:
				if err != nil {
					w.trap.Catch(err)
				}
				w.cancel()
			case <-w.ctx.Done():
				return
			}
		}
	}()

	return
}

func (w *Wrapped) Close() (err error) {
	//w.cancel()
	return
}

func (w *Wrapped) Wait() (err error) {
	w.wg.Wait()
	err = w.trap.Err()
	return
}

