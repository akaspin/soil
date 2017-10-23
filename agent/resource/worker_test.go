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

func TestWorker_Submit(t *testing.T) {
	ctx := context.Background()
	log := logx.GetLog("test")
	cons1 := &bus.DummyConsumer{}

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

	worker := resource.NewWorker(ctx, log, "dummy1", cons1, resource.EvaluatorConfig{}, recovered)

	t.Run("0 configure", func(t *testing.T) {
		worker.Configure(resource.Config{
			Nature: "dummy",
			Kind:   "dummy1",
		})

		time.Sleep(time.Millisecond * 200)
		cons1.AssertMessages(t,
			bus.NewMessage("dummy1", map[string]string{
				"1.res-1.allocated": "true",
				"1.res-1.1":         "2",
				"1.res-1.__values":  "{\"1\":\"2\",\"allocated\":\"true\"}",
				"2.res-2.allocated": "true",
				"2.res-2.__values":  "{\"allocated\":\"true\"}",
			}),
		)
	})
	t.Run("1 remove 2.res-2", func(t *testing.T) {
		worker.Submit("2", nil)
		time.Sleep(time.Millisecond * 200)
		cons1.AssertMessages(t,
			bus.NewMessage("dummy1", map[string]string{
				"1.res-1.allocated": "true",
				"1.res-1.1":         "2",
				"1.res-1.__values":  "{\"1\":\"2\",\"allocated\":\"true\"}",
				"2.res-2.allocated": "true",
				"2.res-2.__values":  "{\"allocated\":\"true\"}",
			}),
			bus.NewMessage("dummy1", map[string]string{
				"1.res-1.allocated": "true",
				"1.res-1.1":         "2",
				"1.res-1.__values":  "{\"1\":\"2\",\"allocated\":\"true\"}",
			}),
		)
	})
	t.Run("2 change 1.res1", func(t *testing.T) {
		worker.Submit("1", []manifest.Resource{
			{
				Kind: "dummy1",
				Name: "res-1",
				Config: map[string]interface{}{
					"1": 1,
				},
			},
		})
		time.Sleep(time.Millisecond * 200)
		cons1.AssertMessages(t,
			bus.NewMessage("dummy1", map[string]string{
				"1.res-1.allocated": "true",
				"1.res-1.1":         "2",
				"1.res-1.__values":  "{\"1\":\"2\",\"allocated\":\"true\"}",
				"2.res-2.allocated": "true",
				"2.res-2.__values":  "{\"allocated\":\"true\"}",
			}),
			bus.NewMessage("dummy1", map[string]string{
				"1.res-1.allocated": "true",
				"1.res-1.1":         "2",
				"1.res-1.__values":  "{\"1\":\"2\",\"allocated\":\"true\"}",
			}),
			bus.NewMessage("dummy1", map[string]string{
				"1.res-1.allocated": "true",
				"1.res-1.1":         "1",
				"1.res-1.__values":  "{\"1\":\"1\",\"allocated\":\"true\"}",
			}),
		)
	})
	t.Run("3 no changes", func(t *testing.T) {
		worker.Submit("1", []manifest.Resource{
			{
				Kind: "dummy1",
				Name: "res-1",
				Config: map[string]interface{}{
					"1": 1,
				},
			},
		})
		time.Sleep(time.Millisecond * 200)
		cons1.AssertMessages(t,
			bus.NewMessage("dummy1", map[string]string{
				"1.res-1.allocated": "true",
				"1.res-1.1":         "2",
				"1.res-1.__values":  "{\"1\":\"2\",\"allocated\":\"true\"}",
				"2.res-2.allocated": "true",
				"2.res-2.__values":  "{\"allocated\":\"true\"}",
			}),
			bus.NewMessage("dummy1", map[string]string{
				"1.res-1.allocated": "true",
				"1.res-1.1":         "2",
				"1.res-1.__values":  "{\"1\":\"2\",\"allocated\":\"true\"}",
			}),
			bus.NewMessage("dummy1", map[string]string{
				"1.res-1.allocated": "true",
				"1.res-1.1":         "1",
				"1.res-1.__values":  "{\"1\":\"1\",\"allocated\":\"true\"}",
			}),
		)
	})
	t.Run("4 add 2", func(t *testing.T) {
		worker.Submit("2", []manifest.Resource{
			{
				Kind:   "dummy1",
				Name:   "res-1",
				Config: map[string]interface{}{},
			},
			{
				Kind:   "dummy1",
				Name:   "res-2",
				Config: map[string]interface{}{},
			},
		})
		time.Sleep(time.Millisecond * 200)
		cons1.AssertMessages(t,
			bus.NewMessage("dummy1", map[string]string{
				"1.res-1.allocated": "true",
				"1.res-1.1":         "2",
				"1.res-1.__values":  "{\"1\":\"2\",\"allocated\":\"true\"}",
				"2.res-2.allocated": "true",
				"2.res-2.__values":  "{\"allocated\":\"true\"}",
			}),
			bus.NewMessage("dummy1", map[string]string{
				"1.res-1.allocated": "true",
				"1.res-1.1":         "2",
				"1.res-1.__values":  "{\"1\":\"2\",\"allocated\":\"true\"}",
			}),
			bus.NewMessage("dummy1", map[string]string{
				"1.res-1.allocated": "true",
				"1.res-1.1":         "1",
				"1.res-1.__values":  "{\"1\":\"1\",\"allocated\":\"true\"}",
			}),
			bus.NewMessage("dummy1", map[string]string{
				"1.res-1.allocated": "true",
				"1.res-1.1":         "1",
				"1.res-1.__values":  "{\"1\":\"1\",\"allocated\":\"true\"}",
				"2.res-1.allocated": "true",
				"2.res-1.__values":  "{\"allocated\":\"true\"}",
				"2.res-2.allocated": "true",
				"2.res-2.__values":  "{\"allocated\":\"true\"}",
			}),
		)
	})
	t.Run("5 reconfigure", func(t *testing.T) {
		worker.Configure(resource.Config{
			Nature: "dummy",
			Kind:   "dummy1",
			Properties: map[string]interface{}{
				"prop1": true,
			},
		})
		time.Sleep(time.Millisecond * 200)
		cons1.AssertMessages(t,
			bus.NewMessage("dummy1", map[string]string{
				"1.res-1.allocated": "true",
				"1.res-1.1":         "2",
				"1.res-1.__values":  "{\"1\":\"2\",\"allocated\":\"true\"}",
				"2.res-2.allocated": "true",
				"2.res-2.__values":  "{\"allocated\":\"true\"}",
			}),
			bus.NewMessage("dummy1", map[string]string{
				"1.res-1.allocated": "true",
				"1.res-1.1":         "2",
				"1.res-1.__values":  "{\"1\":\"2\",\"allocated\":\"true\"}",
			}),
			bus.NewMessage("dummy1", map[string]string{
				"1.res-1.allocated": "true",
				"1.res-1.1":         "1",
				"1.res-1.__values":  "{\"1\":\"1\",\"allocated\":\"true\"}",
			}),
			bus.NewMessage("dummy1", map[string]string{
				"1.res-1.allocated": "true",
				"1.res-1.1":         "1",
				"1.res-1.__values":  "{\"1\":\"1\",\"allocated\":\"true\"}",
				"2.res-1.allocated": "true",
				"2.res-1.__values":  "{\"allocated\":\"true\"}",
				"2.res-2.allocated": "true",
				"2.res-2.__values":  "{\"allocated\":\"true\"}",
			}),
			bus.NewMessage("dummy1", map[string]string{
				"1.res-1.allocated": "true",
				"1.res-1.prop1":     "true",
				"1.res-1.1":         "1",
				"1.res-1.__values":  "{\"1\":\"1\",\"allocated\":\"true\",\"prop1\":\"true\"}",
				"2.res-1.allocated": "true",
				"2.res-1.prop1":     "true",
				"2.res-1.__values":  "{\"allocated\":\"true\",\"prop1\":\"true\"}",
				"2.res-2.allocated": "true",
				"2.res-2.prop1":     "true",
				"2.res-2.__values":  "{\"allocated\":\"true\",\"prop1\":\"true\"}",
			}),
		)
	})
	t.Run("5 reconfigure with equal config", func(t *testing.T) {
		worker.Configure(resource.Config{
			Nature: "dummy",
			Kind:   "dummy1",
			Properties: map[string]interface{}{
				"prop1": true,
			},
		})
		time.Sleep(time.Millisecond * 200)
		cons1.AssertMessages(t,
			bus.NewMessage("dummy1", map[string]string{
				"1.res-1.allocated": "true",
				"1.res-1.1":         "2",
				"1.res-1.__values":  "{\"1\":\"2\",\"allocated\":\"true\"}",
				"2.res-2.allocated": "true",
				"2.res-2.__values":  "{\"allocated\":\"true\"}",
			}),
			bus.NewMessage("dummy1", map[string]string{
				"1.res-1.allocated": "true",
				"1.res-1.1":         "2",
				"1.res-1.__values":  "{\"1\":\"2\",\"allocated\":\"true\"}",
			}),
			bus.NewMessage("dummy1", map[string]string{
				"1.res-1.allocated": "true",
				"1.res-1.1":         "1",
				"1.res-1.__values":  "{\"1\":\"1\",\"allocated\":\"true\"}",
			}),
			bus.NewMessage("dummy1", map[string]string{
				"1.res-1.allocated": "true",
				"1.res-1.1":         "1",
				"1.res-1.__values":  "{\"1\":\"1\",\"allocated\":\"true\"}",
				"2.res-1.allocated": "true",
				"2.res-1.__values":  "{\"allocated\":\"true\"}",
				"2.res-2.allocated": "true",
				"2.res-2.__values":  "{\"allocated\":\"true\"}",
			}),
			bus.NewMessage("dummy1", map[string]string{
				"1.res-1.allocated": "true",
				"1.res-1.prop1":     "true",
				"1.res-1.1":         "1",
				"1.res-1.__values":  "{\"1\":\"1\",\"allocated\":\"true\",\"prop1\":\"true\"}",
				"2.res-1.allocated": "true",
				"2.res-1.prop1":     "true",
				"2.res-1.__values":  "{\"allocated\":\"true\",\"prop1\":\"true\"}",
				"2.res-2.allocated": "true",
				"2.res-2.prop1":     "true",
				"2.res-2.__values":  "{\"allocated\":\"true\",\"prop1\":\"true\"}",
			}),
		)
	})
	t.Run("6 remove allocations", func(t *testing.T) {
		worker.Submit("1", nil)
		worker.Submit("2", nil)
		time.Sleep(time.Millisecond * 200)
		cons1.AssertLastMessage(t,
			bus.NewMessage("dummy1", map[string]string{}),
		)
	})

	worker.Close()
}

func TestWorker_Configure(t *testing.T) {
	t.Run("0 empty", func(t *testing.T) {
		cons := &bus.DummyConsumer{}
		worker := resource.NewWorker(context.Background(), logx.GetLog(""), "dummy1", cons, resource.EvaluatorConfig{}, nil)
		worker.Configure(resource.Config{
			Nature: "dummy",
			Kind:   "dummy1",
		})
		time.Sleep(time.Millisecond * 200)
		cons.AssertMessages(t, bus.NewMessage("dummy1", map[string]string{}))
	})
}
