package supervisor_test

import (
	"context"
	"errors"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestGroup_Cycle(t *testing.T) {
	t.Parallel()
	for i := 0; i < compositeTestIterations; i++ {

		t.Run(`empty owc `+strconv.Itoa(i), func(t *testing.T) {
			waitCh := make(chan struct{})
			sv := supervisor.NewGroup(context.Background())

			assert.NoError(t, sv.Open())

			go func() {
				assert.NoError(t, sv.Wait())
				close(waitCh)
			}()

			assert.NoError(t, sv.Open())
			assert.NoError(t, sv.Close())

			<-waitCh
			assert.NoError(t, sv.Wait())
		})
		t.Run(`close first `+strconv.Itoa(i), func(t *testing.T) {
			waitCh := make(chan struct{})
			c1 := newTestingComponent("1", nil, nil, nil)
			c2 := newTestingComponent("2", nil, nil, nil)
			sv := supervisor.NewGroup(context.Background(), c1, c2)

			assert.NoError(t, sv.Close())
			assert.EqualError(t, sv.Open(), "prematurely closed")
			go func() {
				assert.NoError(t, sv.Wait())
				close(waitCh)
			}()
			<-waitCh
			c1.assertEvents(t)
			c1.assertEvents(t)
		})
		t.Run(`owc `+strconv.Itoa(i), func(t *testing.T) {
			waitCh := make(chan struct{})
			c1 := newTestingComponent("1", nil, nil, nil)
			c2 := newTestingComponent("2", nil, nil, nil)
			c3 := newTestingComponent("3", nil, nil, nil)
			sv := supervisor.NewGroup(context.Background(), c1, c2, c3)

			assert.NoError(t, sv.Open())

			go func() {
				assert.NoError(t, sv.Wait())
				close(waitCh)
			}()

			assert.NoError(t, sv.Open())
			assert.NoError(t, sv.Close())

			<-waitCh
			c1.assertCycle(t)
			c2.assertCycle(t)
			c3.assertCycle(t)
		})
		t.Run(`o-error wc `+strconv.Itoa(i), func(t *testing.T) {
			waitCh := make(chan struct{})
			c1 := newTestingComponent("1", nil, nil, nil)
			c2 := newTestingComponent("2", errors.New("2"), nil, nil)
			c3 := newTestingComponent("3", nil, nil, nil)
			sv := supervisor.NewGroup(context.Background(), c1, c2, c3)

			assert.EqualError(t, sv.Open(), "2")

			go func() {
				assert.NoError(t, sv.Wait())
				close(waitCh)
			}()

			assert.EqualError(t, sv.Open(), "2")
			assert.NoError(t, sv.Close())

			<-waitCh
			c1.assertCycle(t)
			c2.assertEvents(t, "open")
			c3.assertCycle(t)
		})
		t.Run(`ow c-error `+strconv.Itoa(i), func(t *testing.T) {
			waitCh := make(chan struct{})
			c1 := newTestingComponent("1", nil, nil, nil)
			c2 := newTestingComponent("2", nil, errors.New("2"), nil)
			c3 := newTestingComponent("3", nil, nil, nil)
			sv := supervisor.NewGroup(context.Background(), c1, c2, c3)

			assert.NoError(t, sv.Open())

			go func() {
				assert.NoError(t, sv.Wait())
				close(waitCh)
			}()

			assert.NoError(t, sv.Open())
			assert.EqualError(t, sv.Close(), "2")

			<-waitCh
			c1.assertCycle(t)
			c2.assertCycle(t)
			c3.assertCycle(t)
		})
		t.Run(`o w-error c `+strconv.Itoa(i), func(t *testing.T) {
			waitCh := make(chan struct{})
			c1 := newTestingComponent("1", nil, nil, nil)
			c2 := newTestingComponent("2", nil, nil, errors.New("2"))
			c3 := newTestingComponent("3", nil, nil, nil)
			sv := supervisor.NewGroup(context.Background(), c1, c2, c3)

			assert.NoError(t, sv.Open())

			go func() {
				assert.EqualError(t, sv.Wait(), "2")
				close(waitCh)
			}()

			assert.NoError(t, sv.Open())
			assert.NoError(t, sv.Close())
			assert.EqualError(t, sv.Wait(), "2")

			<-waitCh
			c1.assertCycle(t)
			c2.assertCycle(t)
			c3.assertCycle(t)
		})
		t.Run(`exit one `+strconv.Itoa(i), func(t *testing.T) {
			waitCh := make(chan struct{})
			c1 := newTestingComponent("1", nil, nil, nil)
			c2 := newTestingComponent("2", nil, nil, nil)
			c3 := newTestingComponent("3", nil, nil, nil)
			sv := supervisor.NewGroup(context.Background(), c1, c2, c3)

			assert.NoError(t, sv.Open())

			go func() {
				assert.NoError(t, sv.Wait())
				close(waitCh)
			}()

			assert.NoError(t, sv.Open())
			close(c2.closedChan)
			assert.NoError(t, sv.Wait())

			<-waitCh
			c1.assertCycle(t)
			c2.assertEvents(t, "open", "done")
			c3.assertCycle(t)
		})
	}

}
