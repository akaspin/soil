// +build ide test_unit

package estimator_test

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/bus/pipe"
	"github.com/akaspin/soil/agent/resource/estimator"
	"github.com/akaspin/soil/fixture"
	"github.com/akaspin/soil/manifest"
	"testing"
)

func TestRange(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("i%d", i), func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			cons := bus.NewTestingConsumer(ctx)
			r := estimator.NewRange(estimator.GlobalConfig{}, estimator.Config{
				Ctx: ctx,
				Log: logx.GetLog("test"),
				Provider: &allocation.Provider{
					Kind: "range",
					Name: "port",
					Config: map[string]interface{}{
						"min": 8000,
						"max": 8003,
					},
				},
			})
			defer r.Close()
			downstream := pipe.NewLift("test", cons)
			_, _, ch := r.Results()
			go func() {
				for res := range ch {
					downstream.ConsumeMessage(res.Message)
				}
			}()

			t.Run("0 recovered in range", func(t *testing.T) {
				r.Create("8080", &allocation.Resource{
					Request: manifest.Resource{
						Provider: "port",
						Name:     "8080",
					},
					Values: manifest.FlatMap{"value": "8002"},
				})
				fixture.WaitNoError10(t, cons.ExpectLastMessageFn(
					bus.NewMessage("test", map[string]string{
						"8080.allocated": "true",
						"8080.value":     "8002",
					}),
				))
			})
			t.Run("recovered not in range", func(t *testing.T) {
				r.Create("8081", &allocation.Resource{
					Request: manifest.Resource{
						Provider: "port",
						Name:     "8081",
					},
					Values: manifest.FlatMap{"value": "1000"},
				})
				fixture.WaitNoError10(t, cons.ExpectLastMessageFn(
					bus.NewMessage("test", map[string]string{
						"8080.allocated": "true",
						"8080.value":     "8002",
						"8081.allocated": "true",
						"8081.value":     "8000",
					}),
				))
			})
			t.Run("without value", func(t *testing.T) {
				r.Create("8082", &allocation.Resource{
					Request: manifest.Resource{Provider: "port", Name: "8082"},
				})
				fixture.WaitNoError10(t, cons.ExpectLastMessageFn(
					bus.NewMessage("test", map[string]string{
						"8080.allocated": "true",
						"8080.value":     "8002",
						"8081.allocated": "true",
						"8081.value":     "8000",
						"8082.allocated": "true",
						"8082.value":     "8001",
					}),
				))
			})
			t.Run("fill up range", func(t *testing.T) {
				r.Create("8083", &allocation.Resource{
					Request: manifest.Resource{Provider: "port", Name: "8083"},
				})

				fixture.WaitNoError10(t, cons.ExpectLastMessageFn(
					bus.NewMessage("test", map[string]string{
						"8080.allocated": "true",
						"8080.value":     "8002",
						"8081.allocated": "true",
						"8081.value":     "8000",
						"8082.allocated": "true",
						"8082.value":     "8001",
						"8083.allocated": "true",
						"8083.value":     "8003",
					}),
				))
			})
			t.Run("not available", func(t *testing.T) {
				r.Create("failed", &allocation.Resource{
					Request: manifest.Resource{Provider: "port", Name: "failed"},
				})
				fixture.WaitNoError10(t, cons.ExpectLastMessageFn(
					bus.NewMessage("test", map[string]string{
						"8080.allocated":   "true",
						"8080.value":       "8002",
						"8081.allocated":   "true",
						"8081.value":       "8000",
						"8082.allocated":   "true",
						"8082.value":       "8001",
						"8083.allocated":   "true",
						"8083.value":       "8003",
						"failed.allocated": "false",
						"failed.failure":   "not-available",
					}),
				))
			})
			t.Run("0 remove 8082", func(t *testing.T) {
				r.Destroy("8082")
				fixture.WaitNoError10(t, cons.ExpectLastMessageFn(
					bus.NewMessage("test", map[string]string{
						"8080.allocated":   "true",
						"8080.value":       "8002",
						"8081.allocated":   "true",
						"8081.value":       "8000",
						"8083.allocated":   "true",
						"8083.value":       "8003",
						"failed.allocated": "true",
						"failed.value":     "8001",
					}),
				))
			})
		})
	}

}
