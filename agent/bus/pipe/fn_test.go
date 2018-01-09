// +build ide test_unit

package pipe_test

import (
	"context"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/bus/pipe"
	"github.com/akaspin/soil/fixture"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFnPipe_ConsumeMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c1 := bus.NewTestingConsumer(ctx)


	mPipe := pipe.NewFn(func(message bus.Message) (res bus.Message) {
		var chunk map[string]string
		err := message.Payload().Unmarshal(&chunk)
		assert.NoError(t, err)
		delete(chunk, "a")
		res = bus.NewMessage(message.Topic(), chunk)
		return
	}, c1)

	mPipe.ConsumeMessage(bus.NewMessage("test", map[string]string{
		"a": "1",
		"b": "2",
	}))

	fixture.WaitNoError(t, fixture.DefaultWaitConfig(), c1.ExpectMessagesFn(
		bus.NewMessage("test", map[string]string{"b": "2"}),
	))
}
