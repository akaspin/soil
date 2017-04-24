package supervisor

import (
	"context"
	"sync"
)

type link struct {
	ascendant *link
	trap      *Trap
	wg        *sync.WaitGroup

	component Component
	ctx       context.Context
	cancel    context.CancelFunc
}

func (l *link) supervise() {
	ctx, cancel := context.WithCancel(context.Background())

	// set upstream watchdog
	go func() {
		// wait reached
		<-ctx.Done()
		if l.ascendant != nil {
			l.ascendant.cancel()
		}
		l.wg.Done()
	}()

	go func() {
		// external close
		<-l.ctx.Done()
		if err := l.component.Close(); err != nil {
			l.trap.Catch(err)
		}
		l.wg.Done()
	}()

	if err := l.component.Wait(); err != nil {
		l.trap.Catch(err)
	}
	cancel()
}

// Chain composes chain of components. Ascendants always opens before
// descendants and closes after. Chain "Wait" blocs until all chain components
// are closed or error in at least in one components. On error whole chain will
// be closed and "Wait" will return first error.
type Chain struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup

	components []Component
	trap       *Trap
}

func NewChain(ctx context.Context, components ...Component) (c *Chain) {
	c = &Chain{
		wg:         &sync.WaitGroup{},
		components: components,
	}
	c.ctx, c.cancel = context.WithCancel(ctx)
	c.trap = NewTrap(c.cancel)
	return
}

func (c *Chain) Open() (err error) {
	background := context.Background()
	var ascendant *link
	for _, component := range c.components {
		if err = component.Open(); err != nil {
			c.Close()
			return
		}
		c.wg.Add(2)
		l := &link{
			ascendant: ascendant,
			trap:      c.trap,
			wg:        c.wg,
			component: component,
		}
		l.ctx, l.cancel = context.WithCancel(background)
		background = l.ctx
		ascendant = l
		go l.supervise()
	}

	go func() {
		<-c.ctx.Done()
		if ascendant != nil {
			ascendant.cancel()
		}
	}()
	return
}

func (c *Chain) Close() (err error) {
	c.cancel()
	return
}

func (c *Chain) Wait() (err error) {
	c.wg.Wait()
	err = c.trap.Err()
	return
}
