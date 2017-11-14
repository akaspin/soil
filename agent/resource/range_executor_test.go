// +build ide test_unit

package resource_test

import (
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/resource"
	"github.com/akaspin/soil/manifest"
	"testing"
	"time"
	"context"
	"github.com/akaspin/soil/fixture"
)

func TestRangeExecutor_Allocate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cons := bus.NewTestingConsumer(ctx)
	executor := resource.NewRangeExecutor(logx.GetLog("test"), resource.Config{
		Nature: "range",
		Kind:   "port",
		Properties: map[string]interface{}{
			"min": 8000,
			"max": 8003,
		},
	}, cons)

	t.Run("0 recovered in range", func(t *testing.T) {
		executor.Allocate(resource.Alloc{
			PodName: "1",
			Request: manifest.Resource{Kind: "port",
				Name: "8080",
			},
			Values: bus.NewMessage("1.8080", map[string]string{"value": "8002"}),
		})
		fixture.WaitNoError(t, fixture.DefaultWaitConfig(), cons.ExpectMessagesFn(
			bus.NewMessage("1.8080", map[string]string{"allocated": "true", "value": "8002"}),
		))
	})
	t.Run("0 recovered not in range", func(t *testing.T) {
		executor.Allocate(resource.Alloc{
			PodName: "1",
			Request: manifest.Resource{Kind: "port", Name: "8081"},
			Values:  bus.NewMessage("1.8080", map[string]string{"value": "1000"}),
		})
		fixture.WaitNoError(t, fixture.DefaultWaitConfig(), cons.ExpectMessagesFn(
			bus.NewMessage("1.8080", map[string]string{"allocated": "true", "value": "8002"}),
			bus.NewMessage("1.8081", map[string]string{"allocated": "true", "value": "8000"}),
		))
	})
	t.Run("0 allocate", func(t *testing.T) {
		executor.Allocate(resource.Alloc{
			PodName: "1",
			Request: manifest.Resource{Kind: "port", Name: "8082"},
		})

		fixture.WaitNoError(t, fixture.DefaultWaitConfig(), cons.ExpectMessagesFn(
			bus.NewMessage("1.8080", map[string]string{"allocated": "true", "value": "8002"}),
			bus.NewMessage("1.8081", map[string]string{"allocated": "true", "value": "8000"}),
			bus.NewMessage("1.8082", map[string]string{"allocated": "true", "value": "8001"}),
		))
	})
	t.Run("0 not available", func(t *testing.T) {
		executor.Allocate(resource.Alloc{
			PodName: "1",
			Request: manifest.Resource{Kind: "port", Name: "8083"},
		})
		time.Sleep(time.Millisecond * 100)
		executor.Allocate(resource.Alloc{
			PodName: "1",
			Request: manifest.Resource{Kind: "port", Name: "failed"},
		})

		fixture.WaitNoError(t, fixture.DefaultWaitConfig(), cons.ExpectMessagesFn(
			bus.NewMessage("1.8080", map[string]string{"allocated": "true", "value": "8002"}),
			bus.NewMessage("1.8081", map[string]string{"allocated": "true", "value": "8000"}),
			bus.NewMessage("1.8082", map[string]string{"allocated": "true", "value": "8001"}),
			bus.NewMessage("1.8083", map[string]string{"allocated": "true", "value": "8003"}),
			bus.NewMessage("1.failed", map[string]string{"allocated": "false", "failure": "not-available"}),
		))
	})
	t.Run("0 remove 1.8082", func(t *testing.T) {
		executor.Deallocate("1.8082")
		fixture.WaitNoError(t, fixture.DefaultWaitConfig(), cons.ExpectMessagesFn(
			bus.NewMessage("1.8080", map[string]string{"allocated": "true", "value": "8002"}),
			bus.NewMessage("1.8081", map[string]string{"allocated": "true", "value": "8000"}),
			bus.NewMessage("1.8082", map[string]string{"allocated": "true", "value": "8001"}),
			bus.NewMessage("1.8083", map[string]string{"allocated": "true", "value": "8003"}),
			bus.NewMessage("1.failed", map[string]string{"allocated": "false", "failure": "not-available"}),
			bus.NewMessage("1.8082", nil),
			bus.NewMessage("1.failed", map[string]string{"allocated": "true", "value": "8001"}),
		))
	})

}
