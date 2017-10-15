// +build ide test_unit

package resource_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/resource"
	"github.com/akaspin/soil/manifest"
	"testing"
	"time"
)

func TestWorker_Configure(t *testing.T) {
	ctx := context.Background()
	log := logx.GetLog("test")

	recovered := []*resource.Allocation{
		{
			PodName: "1",
			Resource: &allocation.Resource{
				Request: manifest.Resource{
					Kind: "dummy1",
					Name: "res-1",
					Config: map[string]interface{}{
						"1": 2,
					},
				},
				Values: map[string]string{
					"1": "2",
				},
			},
		},
		{
			PodName: "2",
			Resource: &allocation.Resource{
				Request: manifest.Resource{
					Kind: "dummy1",
					Name: "res-2",
					Config: map[string]interface{}{},
				},
				Values: map[string]string{},
			},
		},
	}

	worker := resource.NewWorker(ctx, log, "dummy1", resource.EvaluatorConfig{}, recovered)
	worker.Configure(resource.Config{
		Nature: "dummy",
		Kind:   "dummy1",
		Properties: map[string]interface{}{
			"a": 1,
		},
	})

	worker.Configure(resource.Config{
		Nature: "dummy",
		Kind:   "dummy1",
		Properties: map[string]interface{}{
			"a": 2,
		},
	})

	time.Sleep(time.Millisecond * 1000)
	worker.Close()
}
