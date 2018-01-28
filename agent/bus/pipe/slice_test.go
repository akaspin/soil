// +build ide test_unit

package pipe_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/bus/pipe"
	"github.com/akaspin/soil/fixture"
	"testing"
)

func TestSlicerPipe_ConsumeMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cons := bus.NewTestingConsumer(ctx)
	slicer := pipe.NewSlice(logx.GetLog("test"), cons)
	slicer.ConsumeMessage(bus.NewMessage("1", map[string]interface{}{
		"1": 1,
		"2": 2,
	}))

	fixture.WaitNoErrorT10(t, cons.ExpectMessagesFn(
		bus.NewMessage("1", []int{1, 2}),
	))
}
