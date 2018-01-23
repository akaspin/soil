package supervisor

import (
	"context"
	"sync"
	"sync/atomic"
)

/*
Group supervises Components in parallel. All supervised components are open
and closed concurrently.

Group collects and returns error from corresponding Component methods. If more
than one Components returns errors they will be wrapped in errslice.Error.
*/
type Group struct {
	*composite
	components []Component
}

// NewGroup creates new Group. Provided context manages whole Group. Close
// Context is equivalent to call Group.Close().
func NewGroup(ctx context.Context, components ...Component) (g *Group) {
	g = &Group{
		components: components,
	}
	g.composite = newComposite(ctx, g.build)
	return
}

func (g *Group) build(control *compositeControl) {
	var wg sync.WaitGroup
	wg.Add(len(g.components))
	for _, component := range g.components {
		go func(component Component) {
			defer wg.Done()
			if openErr := component.Open(); openErr != nil {
				control.openError.set(openErr)
				control.cancelFunc()
				return
			}
			// close watchdog
			control.closeWg.Add(1)
			var waitExited uint32
			go func() {
				defer control.closeWg.Done()
				<-control.ctx.Done()
				if atomic.CompareAndSwapUint32(&waitExited, 0, 1) {
					if closeErr := component.Close(); closeErr != nil {
						control.closeError.set(closeErr)
					}
				}
			}()
			// wait watchdog
			control.waitWg.Add(1)
			go func() {
				defer control.waitWg.Done()
				if waitErr := component.Wait(); waitErr != nil {
					control.waitError.set(waitErr)
				}
				atomic.CompareAndSwapUint32(&waitExited, 0, 1)
				control.cancelFunc()
			}()
		}(component)
	}
	wg.Wait()
}
