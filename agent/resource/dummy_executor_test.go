// +build ide test_unit

package resource_test

import (
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/resource"
	"github.com/akaspin/soil/manifest"
	"testing"
	"time"
)

func TestDummyExecutor_Allocate(t *testing.T) {
	cons1 := &bus.DummyConsumer{}
	executor := resource.NewDummyExecutor(logx.GetLog("kind-1"), resource.Config{
		Kind: "kind-1",
	}, cons1)

	executor.Allocate(resource.Alloc{
		PodName: "pod-1",
		Request: manifest.Resource{
			Kind: "kind-1",
			Name: "res-1",
			Config: map[string]interface{}{
				"val1": 1,
				"val2": "ok",
			},
		},
		Values: bus.NewMessage("kind-1.res-1", map[string]string{
			"val1": "2",
		}),
	})

	time.Sleep(time.Millisecond * 100)
	cons1.AssertMessages(t,
		bus.NewMessage("pod-1.res-1", map[string]string{
			"allocated": "true",
			"val1":      "1",
			"val2":      "ok",
		}),
	)

	executor.Deallocate("pod-1.res-1")
	time.Sleep(time.Millisecond * 100)
	cons1.AssertMessages(t,
		bus.NewMessage("pod-1.res-1", map[string]string{
			"allocated": "true",
			"val1":      "1",
			"val2":      "ok",
		}),
		bus.NewMessage("pod-1.res-1", nil),
	)

}
