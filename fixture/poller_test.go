package fixture_test

import (
	"testing"
	"github.com/akaspin/soil/fixture"
	"time"
	"sync/atomic"
)

func TestPollEqual(t *testing.T) {
	t.Run("not equal", func(t *testing.T) {
		t.Skip("should be failed")
		var i int32
		fixture.PollEqual(t, 3, time.Millisecond * 100, func() (res interface{}, err error) {
			res = atomic.AddInt32(&i, 1)
			return
		}, "a")
	})
	t.Run("ok", func(t *testing.T) {
		var i int32
		fixture.PollEqual(t, 3, time.Millisecond * 100, func() (res interface{}, err error) {
			res = atomic.AddInt32(&i, 1)
			return
		}, int32(3))
	})
}