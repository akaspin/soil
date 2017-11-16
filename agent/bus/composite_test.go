// +build ide test_unit

package bus_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/fixture"
	"testing"
)

func TestCompositePipe_ConsumeMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dummy := bus.NewTestingConsumer(ctx)
	pipe := bus.NewCompositePipe("test", logx.GetLog("test"), dummy, "1", "2")

	t.Run("1", func(t *testing.T) {
		pipe.ConsumeMessage(bus.NewMessage("1", map[string]string{
			"1": "1",
		}))
		fixture.WaitNoError(t, fixture.DefaultWaitConfig(), dummy.ExpectMessagesFn(
			bus.NewMessage("test", nil),
		))
	})
	t.Run("2", func(t *testing.T) {
		pipe.ConsumeMessage(bus.NewMessage("2", map[string]string{
			"2": "2",
		}))
		fixture.WaitNoError(t, fixture.DefaultWaitConfig(), dummy.ExpectMessagesFn(
			bus.NewMessage("test", nil),
			bus.NewMessage("test", map[string]string{
				"1.1": "1",
				"2.2": "2",
			}),
		))
	})
	t.Run("2 off", func(t *testing.T) {
		pipe.ConsumeMessage(bus.NewMessage("2", nil))
		fixture.WaitNoError(t, fixture.DefaultWaitConfig(), dummy.ExpectMessagesFn(
			bus.NewMessage("test", nil),
			bus.NewMessage("test", map[string]string{
				"1.1": "1",
				"2.2": "2",
			}),
			bus.NewMessage("test", nil),
		))
	})
	t.Run("2 on", func(t *testing.T) {
		pipe.ConsumeMessage(bus.NewMessage("2", map[string]string{
			"2": "3",
		}))
		fixture.WaitNoError(t, fixture.DefaultWaitConfig(), dummy.ExpectMessagesFn(
			bus.NewMessage("test", nil),
			bus.NewMessage("test", map[string]string{
				"1.1": "1",
				"2.2": "2",
			}),
			bus.NewMessage("test", nil),
			bus.NewMessage("test", map[string]string{
				"1.1": "1",
				"2.2": "3",
			}),
		))
	})
}
