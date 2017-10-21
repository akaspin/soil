// +build ide test_unit

package resource_test

import (
	"github.com/akaspin/soil/agent/resource"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExecutorConfig_IsEqual(t *testing.T) {
	left := resource.ExecutorConfig{
		Nature: "test",
		Kind:   "test",
		Properties: map[string]interface{}{
			"1": 1,
		},
	}
	t.Run("equal", func(t *testing.T) {
		assert.True(t, left.IsEqual(resource.ExecutorConfig{
			Nature: "test",
			Kind:   "test",
			Properties: map[string]interface{}{
				"1": 1,
			},
		}))
	})
	t.Run("equal-pointer", func(t *testing.T) {
		assert.True(t, (&left).IsEqual(resource.ExecutorConfig{
			Nature: "test",
			Kind:   "test",
			Properties: map[string]interface{}{
				"1": 1,
			},
		}))
	})
	t.Run("nature", func(t *testing.T) {
		assert.False(t, (&left).IsEqual(resource.ExecutorConfig{
			Nature: "test1",
			Kind:   "test",
			Properties: map[string]interface{}{
				"1": 1,
			},
		}))
	})
	t.Run("kind", func(t *testing.T) {
		assert.False(t, (&left).IsEqual(resource.ExecutorConfig{
			Nature: "test",
			Kind:   "test1",
			Properties: map[string]interface{}{
				"1": 1,
			},
		}))
	})
	t.Run("properties", func(t *testing.T) {
		assert.False(t, (&left).IsEqual(resource.ExecutorConfig{
			Nature: "test",
			Kind:   "test",
			Properties: map[string]interface{}{
				"1": 2,
			},
		}))
	})

}
