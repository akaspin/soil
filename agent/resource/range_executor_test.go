package resource_test

import (
	"testing"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/resource"
	"github.com/akaspin/logx"
	"github.com/stretchr/testify/assert"
	"github.com/kr/pretty"
	"github.com/akaspin/soil/manifest"
	"time"
)

func TestRangeExecutor_Allocate(t *testing.T) {
	cons := &bus.DummyConsumer{}
	executor, err := resource.NewRangeExecutor(logx.GetLog("test"), resource.ExecutorConfig{
		Nature: "range",
		Kind: "port",
		Properties: map[string]interface{}{
			"min": 8000,
			"max": 9999,
		},
	}, cons)
	assert.NoError(t, err)
	executor.Allocate(resource.Alloc{
		PodName: "1",
		Request: manifest.Resource{
			Kind: "port",
			Name: "8080",
			Config: map[string]interface{}{
				"value": 8090,
			},
		},
		Values: bus.NewMessage("1.8080", map[string]string{
			"value": "8080",
		}),
	})
	executor.Allocate(resource.Alloc{
		PodName: "1",
		Request: manifest.Resource{
			Kind: "port",
			Name: "9090",
			Config: map[string]interface{}{},
		},
		Values: bus.NewMessage("1.8080", nil),
	})
	executor.Allocate(resource.Alloc{
		PodName: "1",
		Request: manifest.Resource{
			Kind: "port",
			Name: "9090",
			Config: map[string]interface{}{},
		},
		Values: bus.NewMessage("1.8080", nil),
	})
	time.Sleep(time.Millisecond * 200)

	executor.Allocate(resource.Alloc{
		PodName: "1",
		Request: manifest.Resource{
			Kind: "port",
			Name: "8081",
			Config: map[string]interface{}{
				"value": 8090,
			},
		},
	})
	time.Sleep(time.Millisecond * 200)
	executor.Deallocate("1.8080")
	time.Sleep(time.Millisecond * 200)
	executor.Allocate(resource.Alloc{
		PodName: "1",
		Request: manifest.Resource{
			Kind: "port",
			Name: "8081",
			Config: map[string]interface{}{
				"value": 8091,
			},
		},
	})
	time.Sleep(time.Millisecond * 200)

	pretty.Log(cons)
}
