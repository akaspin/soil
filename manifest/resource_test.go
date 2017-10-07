// +build ide test_unit

package manifest_test

import (
	"testing"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
)

func TestResource(t *testing.T) {
	res := manifest.Resource{
		Type: "port",
		Name: "8080",
		Required: true,
	}
	podName := "test"

	t.Run("Id", func(t *testing.T) {
		assert.Equal(t, "test.8080", res.GetId(podName))
	})
	t.Run("request constraint", func(t *testing.T) {
		assert.Equal(t, manifest.Constraint{
			"${__resource.request.type.port}":"true",
		}, res.GetRequestConstraint())
	})
	t.Run("allocation constraint", func(t *testing.T) {
		assert.Equal(t, manifest.Constraint{
			"${resource.port.test.8080.allocated}":"true",
		}, res.GetAllocationConstraint(podName))
	})
	t.Run("values key", func(t *testing.T) {
		assert.Equal(t, "__resource.values.port.test.8080", res.GetValuesKey(podName))
	})
}
