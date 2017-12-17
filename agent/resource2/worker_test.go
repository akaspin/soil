// +build ide test_unit

package resource2_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/resource2"
	"github.com/akaspin/soil/fixture"
	"github.com/akaspin/soil/manifest"
	"testing"
)

func TestWorker_Submit(t *testing.T) {
	t.Skip()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log := logx.GetLog("test")
	cons1 := bus.NewTestingConsumer(ctx)

	recovered := []resource2.Alloc{
		{
			PodName: "1",
			Request: manifest.Resource{
				Provider: "dummy1",
				Name:     "res-1",
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
				Provider: "dummy1",
				Name:     "res-2",
				Config:   map[string]interface{}{},
			},
			Values: bus.NewMessage("2.res-2", map[string]string{
				"1": "2",
			}),
		},
	}

	worker := resource2.NewWorker(ctx, log, "dummy1", cons1, resource2.EvaluatorConfig{}, recovered)

	waitConfig := fixture.DefaultWaitConfig()
	t.Run("0 configure", func(t *testing.T) {
		worker.Configure(resource2.Config{
			Nature: "dummy",
			Kind:   "dummy1",
		})

		fixture.WaitNoError(t, waitConfig, cons1.ExpectMessagesFn(
			bus.NewMessage("dummy1", map[string]string{
				"1.res-1.allocated": "true",
				"1.res-1.1":         "2",
				"1.res-1.__values":  "{\"1\":\"2\",\"allocated\":\"true\"}",
				"2.res-2.allocated": "true",
				"2.res-2.__values":  "{\"allocated\":\"true\"}",
			}),
		))
	})
	t.Run("1 remove 2.res-2", func(t *testing.T) {
		worker.Submit("2", nil)
		fixture.WaitNoError(t, waitConfig, cons1.ExpectMessagesFn(
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
		))
	})
	t.Run("2 change 1.res1", func(t *testing.T) {
		worker.Submit("1", []manifest.Resource{
			{
				Provider: "dummy1",
				Name:     "res-1",
				Config: map[string]interface{}{
					"1": 1,
				},
			},
		})

		fixture.WaitNoError(t, waitConfig, cons1.ExpectMessagesFn(
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
		))
	})
	t.Run("3 no changes", func(t *testing.T) {
		worker.Submit("1", []manifest.Resource{
			{
				Provider: "dummy1",
				Name:     "res-1",
				Config: map[string]interface{}{
					"1": 1,
				},
			},
		})

		fixture.WaitNoError(t, waitConfig, cons1.ExpectMessagesFn(
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
		))
	})
	t.Run("4 add 2", func(t *testing.T) {
		worker.Submit("2", []manifest.Resource{
			{
				Provider: "dummy1",
				Name:     "res-1",
				Config:   map[string]interface{}{},
			},
			{
				Provider: "dummy1",
				Name:     "res-2",
				Config:   map[string]interface{}{},
			},
		})

		fixture.WaitNoError(t, waitConfig, cons1.ExpectMessagesFn(
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
		))
	})
	t.Run("5 reconfigure", func(t *testing.T) {
		worker.Configure(resource2.Config{
			Nature: "dummy",
			Kind:   "dummy1",
			Properties: map[string]interface{}{
				"prop1": true,
			},
		})

		fixture.WaitNoError(t, waitConfig, cons1.ExpectMessagesFn(
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
		))
	})
	t.Run("5 reconfigure with equal config", func(t *testing.T) {
		worker.Configure(resource2.Config{
			Nature: "dummy",
			Kind:   "dummy1",
			Properties: map[string]interface{}{
				"prop1": true,
			},
		})
		fixture.WaitNoError(t, waitConfig, cons1.ExpectMessagesFn(
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
		))
	})
	t.Run("6 remove allocations", func(t *testing.T) {
		worker.Submit("1", nil)
		worker.Submit("2", nil)

		fixture.WaitNoError(t, waitConfig, cons1.ExpectLastMessageFn(
			bus.NewMessage("dummy1", map[string]string{}),
		))
	})

	worker.Close()
}

func TestWorker_Configure(t *testing.T) {
	t.Skip()
	t.Run("0 empty", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cons := bus.NewTestingConsumer(ctx)
		worker := resource2.NewWorker(context.Background(), logx.GetLog(""), "dummy1", cons, resource2.EvaluatorConfig{}, nil)
		worker.Configure(resource2.Config{
			Nature: "dummy",
			Kind:   "dummy1",
		})
		fixture.WaitNoError(t, fixture.DefaultWaitConfig(), cons.ExpectMessagesFn(
			bus.NewMessage("dummy1", map[string]string{}),
		))
	})
}
