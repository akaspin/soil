// +build ide test_unit

package bus_test

import (
	"github.com/akaspin/soil/agent/bus"
	"testing"
)

func TestMapUpstream_Set(t *testing.T) {
	cons2 := &bus.TestConsumer{}

	prod := bus.NewMapUpstream("meta", cons2)
	prod.Set(map[string]string{
		"1": "1",
	})
	cons2.AssertPayloads(t, []map[string]string{
		{"1": "1"},
	})
	prod.Set(map[string]string{
		"2": "2",
	})
	cons2.AssertPayloads(t, []map[string]string{
		{"1": "1"},
		{"1": "1", "2": "2"},
	})
	prod.Set(map[string]string{
		"2": "2",
	})
	cons2.AssertPayloads(t, []map[string]string{
		{"1": "1"},
		{"1": "1", "2": "2"},
	})
	prod.Delete("2", "3")
	cons2.AssertPayloads(t, []map[string]string{
		{"1": "1"},
		{"1": "1", "2": "2"},
		{"1": "1"},
	})
	prod.Delete("2", "3")
	cons2.AssertPayloads(t, []map[string]string{
		{"1": "1"},
		{"1": "1", "2": "2"},
		{"1": "1"},
	})
}

func TestFlatMap_Set(t *testing.T) {
	t.Run("strict", func(t *testing.T) {
		cons2 := &bus.TestConsumer{}

		prod := bus.NewStrictMapUpstream("meta", cons2)
		prod.Set(map[string]string{
			"1": "1",
		})
		cons2.AssertPayloads(t, []map[string]string{
			{"1": "1"},
		})
		prod.Set(map[string]string{
			"2": "2",
		})
		cons2.AssertPayloads(t, []map[string]string{
			{"1": "1"},
			{"2": "2"},
		})
		prod.Set(map[string]string{
			"2": "2",
		})
		cons2.AssertPayloads(t, []map[string]string{
			{"1": "1"},
			{"2": "2"},
		})
		prod.Set(map[string]string{})
		cons2.AssertPayloads(t, []map[string]string{
			{"1": "1"},
			{"2": "2"},
			{},
		})

	})
}
