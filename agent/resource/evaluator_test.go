// +build ide test_unit

package resource_test

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/resource"
	"github.com/akaspin/soil/fixture"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEvaluator_GetConstraint(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	upstream := bus.NewTestingConsumer(ctx)
	downstream := bus.NewTestingConsumer(ctx)
	evaluator := resource.NewEvaluator(ctx, logx.GetLog("test"), upstream, downstream, nil)
	assert.NoError(t, evaluator.Open())

	t.Run(`with resources`, func(t *testing.T) {
		assert.Equal(t,
			manifest.Constraint{
				"1": "2",
				"${provider.test.1.allocated}": "true",
				"${provider.test.3.allocated}": "true",
			},
			evaluator.GetConstraint(&manifest.Pod{
				Constraint: manifest.Constraint{
					"1": "2",
				},
				Resources: manifest.Resources{
					{
						Name:     "1",
						Provider: "test.1",
					},
					{
						Name:     "2",
						Provider: "test.1",
					},
					{
						Name:     "3",
						Provider: "test.3",
					},
				},
			}))
	})
	t.Run(`no resources`, func(t *testing.T) {
		assert.Equal(t,
			manifest.Constraint{
				"resource.allocate": "false",
			},
			evaluator.GetConstraint(&manifest.Pod{
				Constraint: manifest.Constraint{
					"1": "2",
				},
			}))
	})
}

func TestNewEvaluator(t *testing.T) {
	for i := 0; i < 5; i++ {
		t.Run(fmt.Sprintf("i_%d", i), func(t *testing.T) {
			t.Parallel()
			dirty := allocation.PodSlice{
				{
					Header: allocation.Header{
						Name: "pod-0",
					},
					Resources: allocation.ResourceSlice{
						{
							Request: manifest.Resource{
								Name:     "0",
								Provider: "res-0.0",
							},
							Values: manifest.FlatMap{
								"allocated": "true",
								"value":     "8000",
							},
						},
						{
							Request: manifest.Resource{
								Name:     "1",
								Provider: "res-0.1",
							},
							Values: manifest.FlatMap{
								"allocated": "true",
								"value":     "8000",
							},
						},
					},
				},
				{
					Header: allocation.Header{
						Name: "pod-1",
					},
					Resources: allocation.ResourceSlice{
						{
							Request: manifest.Resource{
								Name:     "0",
								Provider: "res-0.1",
							},
							Values: manifest.FlatMap{
								"allocated": "true",
								"value":     "8001",
							},
						},
						{
							Request: manifest.Resource{
								Name:     "1",
								Provider: "res-0.2",
							},
							Values: manifest.FlatMap{
								"allocated": "true",
								"value":     "8000",
							},
						},
					},
				},
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			upstream := bus.NewTestingConsumer(ctx)
			downstream := bus.NewTestingConsumer(ctx)
			evaluator := resource.NewEvaluator(ctx, logx.GetLog("test"), upstream, downstream, dirty)
			assert.NoError(t, evaluator.Open())

			t.Run(`recovery`, func(t *testing.T) {
				fixture.WaitNoErrorT10(t, upstream.ExpectMessagesFn(
					bus.NewMessage("provider", map[string]string{
						"res-0.0.allocated": "true",
						"res-0.0.kind":      "blackhole",
						"res-0.1.allocated": "true",
						"res-0.1.kind":      "blackhole",
						"res-0.2.allocated": "true",
						"res-0.2.kind":      "blackhole",
					}),
				))
				fixture.WaitNoErrorT10(t, downstream.ExpectMessagesFn(
					bus.NewMessage("resource", map[string]string{
						"pod-0.0.allocated": "true",
						"pod-0.0.provider":  "res-0.0",
						"pod-0.0.value":     "8000",
						"pod-0.0.__values":  "{\"allocated\":\"true\",\"value\":\"8000\"}",
						"pod-0.1.allocated": "true",
						"pod-0.1.provider":  "res-0.1",
						"pod-0.1.value":     "8000",
						"pod-0.1.__values":  "{\"allocated\":\"true\",\"value\":\"8000\"}",
						"pod-1.0.allocated": "true",
						"pod-1.0.provider":  "res-0.1",
						"pod-1.0.value":     "8001",
						"pod-1.0.__values":  "{\"allocated\":\"true\",\"value\":\"8001\"}",
						"pod-1.1.allocated": "true",
						"pod-1.1.provider":  "res-0.2",
						"pod-1.1.value":     "8000",
						"pod-1.1.__values":  "{\"allocated\":\"true\",\"value\":\"8000\"}",
					}),
				))
			})
			t.Run(`recover provider res-0.0`, func(t *testing.T) {
				evaluator.CreateProvider("res-0.0", &allocation.Provider{
					Name: "0",
					Kind: "range",
					Config: map[string]interface{}{
						"min": 3000,
						"max": 3002,
					},
				})
				fixture.WaitNoErrorT10(t, upstream.ExpectLastMessageFn(
					bus.NewMessage("provider", map[string]string{
						"res-0.0.allocated": "true",
						"res-0.0.kind":      "range",
						"res-0.1.allocated": "true",
						"res-0.1.kind":      "blackhole",
						"res-0.2.allocated": "true",
						"res-0.2.kind":      "blackhole",
					}),
				))
				fixture.WaitNoErrorT10(t, downstream.ExpectLastMessageFn(
					bus.NewMessage("resource", map[string]string{
						"pod-0.0.allocated": "true",
						"pod-0.0.value":     "3000",
						"pod-0.0.provider":  "res-0.0",
						"pod-0.0.__values":  "{\"allocated\":\"true\",\"value\":\"3000\"}",
						"pod-0.1.allocated": "true",
						"pod-0.1.provider":  "res-0.1",
						"pod-0.1.value":     "8000",
						"pod-0.1.__values":  "{\"allocated\":\"true\",\"value\":\"8000\"}",
						"pod-1.0.allocated": "true",
						"pod-1.0.provider":  "res-0.1",
						"pod-1.0.value":     "8001",
						"pod-1.0.__values":  "{\"allocated\":\"true\",\"value\":\"8001\"}",
						"pod-1.1.allocated": "true",
						"pod-1.1.provider":  "res-0.2",
						"pod-1.1.value":     "8000",
						"pod-1.1.__values":  "{\"allocated\":\"true\",\"value\":\"8000\"}",
					}),
				))
			})
			t.Run(`deallocate pod-1`, func(t *testing.T) {
				evaluator.Deallocate("pod-1")
				fixture.WaitNoErrorT10(t, downstream.ExpectLastMessageFn(
					bus.NewMessage("resource", map[string]string{
						"pod-0.0.allocated": "true",
						"pod-0.0.value":     "3000",
						"pod-0.0.provider":  "res-0.0",
						"pod-0.0.__values":  "{\"allocated\":\"true\",\"value\":\"3000\"}",
						"pod-0.1.allocated": "true",
						"pod-0.1.provider":  "res-0.1",
						"pod-0.1.value":     "8000",
						"pod-0.1.__values":  "{\"allocated\":\"true\",\"value\":\"8000\"}",
					}),
				))
			})
			t.Run(`allocate pod-2`, func(t *testing.T) {
				evaluator.Allocate(&manifest.Pod{
					Name: "pod-2",
					Resources: manifest.Resources{
						{
							Name:     "0",
							Provider: "res-0.0",
						},
					},
				}, map[string]string{})
				fixture.WaitNoErrorT10(t, downstream.ExpectLastMessageFn(
					bus.NewMessage("resource", map[string]string{
						"pod-0.0.allocated": "true",
						"pod-0.0.value":     "3000",
						"pod-0.0.provider":  "res-0.0",
						"pod-0.0.__values":  "{\"allocated\":\"true\",\"value\":\"3000\"}",
						"pod-0.1.allocated": "true",
						"pod-0.1.provider":  "res-0.1",
						"pod-0.1.value":     "8000",
						"pod-0.1.__values":  "{\"allocated\":\"true\",\"value\":\"8000\"}",
						"pod-2.0.allocated": "true",
						"pod-2.0.value":     "3001",
						"pod-2.0.provider":  "res-0.0",
						"pod-2.0.__values":  "{\"allocated\":\"true\",\"value\":\"3001\"}",
					}),
				))
			})
			t.Run(`destroy provider res-0.1`, func(t *testing.T) {
				evaluator.DestroyProvider("res-0.1")
				fixture.WaitNoErrorT10(t, downstream.ExpectLastMessageFn(
					bus.NewMessage("resource", map[string]string{
						"pod-0.0.allocated": "true",
						"pod-0.0.value":     "3000",
						"pod-0.0.provider":  "res-0.0",
						"pod-0.0.__values":  "{\"allocated\":\"true\",\"value\":\"3000\"}",
						"pod-2.0.allocated": "true",
						"pod-2.0.value":     "3001",
						"pod-2.0.provider":  "res-0.0",
						"pod-2.0.__values":  "{\"allocated\":\"true\",\"value\":\"3001\"}",
					}),
				))
			})

			evaluator.Close()
			evaluator.Wait()
		})
	}

}
