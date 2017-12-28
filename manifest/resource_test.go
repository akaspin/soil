// +build ide test_unit

package manifest_test

import (
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResource(t *testing.T) {
	res := manifest.Resource{
		Provider: "port",
		Name:     "8080",
		Config: map[string]interface{}{
			"a": 1,
		},
	}
	podName := "test"

	t.Run("Id", func(t *testing.T) {
		assert.Equal(t, "test.8080", res.GetID(podName))
	})
	t.Run("request constraint", func(t *testing.T) {
		assert.Equal(t, manifest.Constraint{
			"${resource.request.port.allow}": "true",
		}, res.GetRequestConstraint())
	})
	t.Run("allocation constraint", func(t *testing.T) {
		assert.Equal(t, manifest.Constraint{
			"${resource.port.test.8080.allocated}": "true",
		}, res.GetAllocationConstraint(podName))
	})
	t.Run("values key", func(t *testing.T) {
		assert.Equal(t, "resource.port.test.8080.__values", res.GetValuesKey(podName))
	})
	t.Run("clone", func(t *testing.T) {
		res1 := res.Clone()
		assert.Equal(t, res, res1)
		res1.Config["a"] = 2
		assert.NotEqual(t, res, res1)
	})
}
