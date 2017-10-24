// +build ide test_unit

package bus_test

import (
	"github.com/akaspin/soil/agent/bus"
	"testing"
)

func TestStorage_ConsumeMessage(t *testing.T) {
	cons2 := &bus.DummyConsumer{}

	prod := bus.NewStorage("meta", cons2)
	prod.ConsumeMessage(bus.NewMessage("1", map[string]string{
		"1": "1",
	}))
	cons2.AssertMessages(t,
		bus.NewMessage("meta", map[string]string{
			"1.1": "1",
		}),
	)
	prod.ConsumeMessage(bus.NewMessage("2", map[string]string{
		"2": "2",
	}))
	cons2.AssertMessages(t,
		bus.NewMessage("meta", map[string]string{
			"1.1": "1",
		}),
		bus.NewMessage("meta", map[string]string{
			"1.1": "1",
			"2.2": "2",
		}),
	)
	prod.ConsumeMessage(bus.NewMessage("2", nil))
	cons2.AssertMessages(t,
		bus.NewMessage("meta", map[string]string{
			"1.1": "1",
		}),
		bus.NewMessage("meta", map[string]string{
			"1.1": "1",
			"2.2": "2",
		}),
		bus.NewMessage("meta", map[string]string{
			"1.1": "1",
		}),
	)
}
