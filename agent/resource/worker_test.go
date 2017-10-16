// +build ide test_unit

package resource_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/resource"
	"github.com/akaspin/soil/manifest"
	"testing"
	"time"
)

func TestWorker_Configure(t *testing.T) {
	ctx := context.Background()
	log := logx.GetLog("test")

	recovered := []resource.Alloc{
		{
			PodName: "1",
			Request: manifest.Resource{
				Kind: "dummy1",
				Name: "res-1",
				Config: map[string]interface{}{
					"1": 2,
				},
			},
			Values: bus.NewMessage("1.res-1", map[string]string{
				"1": "2",
			}),
		},
		{
			PodName: "2",
			Request: manifest.Resource{
				Kind:   "dummy1",
				Name:   "res-2",
				Config: map[string]interface{}{},
			},
			Values: bus.NewMessage("2.res-2", map[string]string{
				"1": "2",
			}),
		},
	}

	worker := resource.NewWorker(ctx, log, "dummy1", resource.EvaluatorConfig{}, recovered)
	worker.Configure(resource.ExecutorConfig{
		Nature: "dummy",
		Kind:   "dummy1",
		Properties: map[string]interface{}{
			"a": 1,
		},
	})

	worker.Configure(resource.ExecutorConfig{
		Nature: "dummy",
		Kind:   "dummy1",
		Properties: map[string]interface{}{
			"a": 2,
		},
	})

	time.Sleep(time.Millisecond * 200)
	worker.Close()
}
