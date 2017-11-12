// +build ide test_unit

package bus_test

import (
	"github.com/akaspin/soil/agent/bus"
	"testing"
)

func TestDivertPipe_Divert(t *testing.T) {
	dummy := &bus.TestingConsumer{}
	pipe := bus.NewDivertPipe("drain", dummy, bus.NewMessage("drain", map[string]string{
		"drain": "true",
	}))

	t.Run("initial message", func(t *testing.T) {
		pipe.ConsumeMessage(bus.NewMessage("test", map[string]string{"test": "1"}))
		dummy.AssertMessages(t, bus.NewMessage("test", map[string]string{"test": "1"}))
	})
	t.Run("drain on", func(t *testing.T) {
		pipe.Divert(true)
		dummy.AssertMessages(t,
			bus.NewMessage("test", map[string]string{"test": "1"}),
			bus.NewMessage("drain", map[string]string{"drain": "true"}),
		)
	})
	t.Run("message in drain mode", func(t *testing.T) {
		pipe.Divert(true)
		pipe.ConsumeMessage(bus.NewMessage("test", map[string]string{"test": "2"}))
		dummy.AssertMessages(t,
			bus.NewMessage("test", map[string]string{"test": "1"}),
			bus.NewMessage("drain", map[string]string{"drain": "true"}),
		)
	})
	t.Run("drain off", func(t *testing.T) {
		pipe.Divert(false)
		dummy.AssertMessages(t,
			bus.NewMessage("test", map[string]string{"test": "1"}),
			bus.NewMessage("drain", map[string]string{"drain": "true"}),
			bus.NewMessage("test", map[string]string{"test": "2"}),
		)
	})
}
