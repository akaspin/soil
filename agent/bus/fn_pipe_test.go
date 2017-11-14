// +build ide test_unit

package bus_test

import (
	"github.com/akaspin/soil/agent/bus"
	"github.com/stretchr/testify/assert"
	"testing"
	"context"
	"github.com/akaspin/soil/fixture"
)

func TestFnPipe_ConsumeMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c1 := bus.NewTestingConsumer(ctx)
	c2 := bus.NewTestingConsumer(ctx)

	pipe := bus.NewFnPipe(func(message bus.Message) (res bus.Message) {
		var chunk map[string]string
		err := message.Payload().Unmarshal(&chunk)
		assert.NoError(t, err)
		delete(chunk, "a")
		res = bus.NewMessage(message.GetID(), chunk)
		return
	}, c1, c2)

	pipe.ConsumeMessage(bus.NewMessage("test", map[string]string{
		"a": "1",
		"b": "2",
	}))

	fixture.WaitNoError(t, fixture.DefaultWaitConfig(), c1.ExpectMessagesFn(
		bus.NewMessage("test", map[string]string{"b": "2"}),
	))
	fixture.WaitNoError(t, fixture.DefaultWaitConfig(), c2.ExpectMessagesFn(
		bus.NewMessage("test", map[string]string{"b": "2"}),
	))
}
