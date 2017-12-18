// +build ide test_unit

package estimator_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/resource/estimator"
	"github.com/akaspin/soil/fixture"
	"github.com/akaspin/soil/manifest"
	"testing"
)

func TestRange_Allocate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cons := bus.NewTestingConsumer(ctx)
	executor := estimator.NewRange(ctx, logx.GetLog("test"), &allocation.Provider{
		Kind: "range",
		Name: "port",
		Config: map[string]interface{}{
			"min": 8000,
			"max": 8003,
		},
	}, cons)

	t.Run("0 recovered in range", func(t *testing.T) {
		executor.Allocate(&allocation.Resource{
			Request: manifest.Resource{
				Provider: "port",
				Name:     "8080",
			},
			Values: manifest.Environment{"value": "8002"},
		})
		fixture.WaitNoError10(t, cons.ExpectMessagesFn(
			bus.NewMessage("8080", map[string]string{
				"allocated": "true",
				"value":     "8002",
			}),
		))
	})
	t.Run("recovered not in range", func(t *testing.T) {
		executor.Allocate(&allocation.Resource{
			Request: manifest.Resource{
				Provider: "port",
				Name:     "8081",
			},
			Values: manifest.Environment{"value": "1000"},
		})
		fixture.WaitNoError10(t, cons.ExpectMessagesFn(
			bus.NewMessage("8080", map[string]string{"allocated": "true", "value": "8002"}),
			bus.NewMessage("8081", map[string]string{"allocated": "true", "value": "8000"}),
		))
	})
	t.Run("without value", func(t *testing.T) {
		executor.Allocate(&allocation.Resource{
			Request: manifest.Resource{Provider: "port", Name: "8082"},
		})

		fixture.WaitNoError10(t, cons.ExpectMessagesFn(
			bus.NewMessage("8080", map[string]string{"allocated": "true", "value": "8002"}),
			bus.NewMessage("8081", map[string]string{"allocated": "true", "value": "8000"}),
			bus.NewMessage("8082", map[string]string{"allocated": "true", "value": "8001"}),
		))
	})
	t.Run("fill up range", func(t *testing.T) {
		executor.Allocate(&allocation.Resource{
			Request: manifest.Resource{Provider: "port", Name: "8083"},
		})

		fixture.WaitNoError10(t, cons.ExpectMessagesFn(
			bus.NewMessage("8080", map[string]string{"allocated": "true", "value": "8002"}),
			bus.NewMessage("8081", map[string]string{"allocated": "true", "value": "8000"}),
			bus.NewMessage("8082", map[string]string{"allocated": "true", "value": "8001"}),
			bus.NewMessage("8083", map[string]string{"allocated": "true", "value": "8003"}),
		))
	})
	t.Run("not available", func(t *testing.T) {
		executor.Allocate(&allocation.Resource{
			Request: manifest.Resource{Provider: "port", Name: "failed"},
		})

		fixture.WaitNoError10(t, cons.ExpectMessagesFn(
			bus.NewMessage("8080", map[string]string{"allocated": "true", "value": "8002"}),
			bus.NewMessage("8081", map[string]string{"allocated": "true", "value": "8000"}),
			bus.NewMessage("8082", map[string]string{"allocated": "true", "value": "8001"}),
			bus.NewMessage("8083", map[string]string{"allocated": "true", "value": "8003"}),
			bus.NewMessage("failed", map[string]string{"allocated": "false", "failure": "not-available"}),
		))
	})
	t.Run("0 remove 8082", func(t *testing.T) {
		executor.Deallocate("8082")
		fixture.WaitNoError10(t, cons.ExpectMessagesFn(
			bus.NewMessage("8080", map[string]string{"allocated": "true", "value": "8002"}),
			bus.NewMessage("8081", map[string]string{"allocated": "true", "value": "8000"}),
			bus.NewMessage("8082", map[string]string{"allocated": "true", "value": "8001"}),
			bus.NewMessage("8083", map[string]string{"allocated": "true", "value": "8003"}),
			bus.NewMessage("failed", map[string]string{"allocated": "false", "failure": "not-available"}),
			bus.NewMessage("8082", nil),
			bus.NewMessage("failed", map[string]string{"allocated": "true", "value": "8001"}),
		))
	})
}
