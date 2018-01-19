package supervisor_test

import (
	"context"
	"errors"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestChain_Cycle(t *testing.T) {
	t.Parallel()
	for i := 0; i < compositeTestIterations; i++ {
		t.Run(`empty owc `+strconv.Itoa(i), func(t *testing.T) {
			waitCh := make(chan struct{})
			sv := supervisor.NewChain(context.Background())

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
			sv := supervisor.NewChain(context.Background(), c1, c2)

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
		t.Run("o w c "+strconv.Itoa(i), func(t *testing.T) {
			waitCh := make(chan struct{})
			c1 := newTestingComponent("1", nil, nil, nil)
			c2 := newTestingComponent("2", nil, nil, nil)
			c3 := newTestingComponent("3", nil, nil, nil)
			watcher := newTestingWatcher(9, c1, c2, c3)

			sv := supervisor.NewChain(context.Background(), c1, c2, c3)

			assert.NoError(t, sv.Open())
			go func() {
				assert.NoError(t, sv.Wait())
				close(waitCh)
			}()

			assert.NoError(t, sv.Close())
			assert.NoError(t, sv.Open())

			<-waitCh
			assert.NoError(t, sv.Wait())
			c1.assertCycle(t)
			c2.assertCycle(t)
			c3.assertCycle(t)

			watcher.wg.Wait()
			assert.Equal(t, []string{
				"1-open", "2-open", "3-open",
				"3-close", "3-done",
				"2-close", "2-done",
				"1-close", "1-done",
			}, watcher.res)
		})
		t.Run(`o-error w c `+strconv.Itoa(i), func(t *testing.T) {
			waitCh := make(chan struct{})
			c1 := newTestingComponent("1", nil, nil, nil)
			c2 := newTestingComponent("2", errors.New("2"), nil, nil)
			c3 := newTestingComponent("3", nil, nil, nil)
			watcher := newTestingWatcher(4, c1, c2, c3)

			sv := supervisor.NewChain(context.Background(), c1, c2, c3)

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
			c3.assertEvents(t)

			watcher.wg.Wait()
			assert.Equal(t, []string{
				"1-open", "2-open",
				"1-close",
				"1-done",
			}, watcher.res)
		})
		t.Run(`o w c-error `+strconv.Itoa(i), func(t *testing.T) {
			waitCh := make(chan struct{})
			c1 := newTestingComponent("1", nil, nil, nil)
			c2 := newTestingComponent("2", nil, errors.New("2"), nil)
			c3 := newTestingComponent("3", nil, nil, nil)
			watcher := newTestingWatcher(9, c1, c2, c3)

			sv := supervisor.NewChain(context.Background(), c1, c2, c3)

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

			watcher.wg.Wait()
			assert.Equal(t, []string{
				"1-open", "2-open", "3-open",
				"3-close", "3-done",
				"2-close", "2-done",
				"1-close", "1-done",
			}, watcher.res)
		})
		t.Run(`o w-error c `+strconv.Itoa(i), func(t *testing.T) {
			waitCh := make(chan struct{})
			c1 := newTestingComponent("1", nil, nil, nil)
			c2 := newTestingComponent("2", nil, nil, errors.New("2"))
			c3 := newTestingComponent("3", nil, nil, nil)
			watcher := newTestingWatcher(9, c1, c2, c3)

			sv := supervisor.NewChain(context.Background(), c1, c2, c3)

			assert.NoError(t, sv.Open())

			go func() {
				assert.EqualError(t, sv.Wait(), "2")
				close(waitCh)
			}()

			assert.NoError(t, sv.Open())
			assert.NoError(t, sv.Close())

			<-waitCh
			c1.assertCycle(t)
			c2.assertCycle(t)
			c3.assertCycle(t)
			assert.EqualError(t, sv.Wait(), "2")

			watcher.wg.Wait()
			assert.Equal(t, []string{
				"1-open", "2-open", "3-open",
				"3-close", "3-done",
				"2-close", "2-done",
				"1-close", "1-done",
			}, watcher.res)
		})
		t.Run(`exit one `+strconv.Itoa(i), func(t *testing.T) {
			waitCh := make(chan struct{})
			c1 := newTestingComponent("1", nil, nil, nil)
			c2 := newTestingComponent("2", nil, nil, nil)
			c3 := newTestingComponent("3", nil, nil, nil)
			watcher := newTestingWatcher(8, c1, c2, c3)

			sv := supervisor.NewChain(context.Background(), c1, c2, c3)

			assert.NoError(t, sv.Open())

			go func() {
				assert.NoError(t, sv.Wait())
				close(waitCh)
			}()

			assert.NoError(t, sv.Open())
			close(c2.closedChan)

			<-waitCh
			c1.assertCycle(t)
			c2.assertEvents(t, "open", "done")
			c3.assertCycle(t)
			assert.NoError(t, sv.Wait())

			watcher.wg.Wait()
			assert.Equal(t, []string{
				"1-open", "2-open", "3-open",
				"2-done",
				"3-close", "3-done",
				"1-close", "1-done",
			}, watcher.res)
		})
	}
}
