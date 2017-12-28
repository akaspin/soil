// +build ide test_unit

package resource_test

import (
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/resource"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPlan(t *testing.T) {
	left := allocation.ResourceSlice{
		{
			Request: manifest.Resource{
				Name:     "1",
				Provider: "test",
			},
		},
		{
			Request: manifest.Resource{
				Name:     "2",
				Provider: "test",
			},
		},
	}
	t.Run(`create all`, func(t *testing.T) {
		c, u, d := resource.Plan(nil, left)
		assert.Equal(t, left, c)
		assert.Empty(t, u)
		assert.Empty(t, d)
	})
	t.Run(`destroy all`, func(t *testing.T) {
		c, u, d := resource.Plan(left, nil)
		assert.Empty(t, c)
		assert.Empty(t, u)
		assert.Equal(t, left, d)
	})
	t.Run(`change c3 u2 d1`, func(t *testing.T) {
		c, u, d := resource.Plan(left, allocation.ResourceSlice{
			{
				Request: manifest.Resource{
					Name:     "2",
					Provider: "test1",
				},
			},
			{
				Request: manifest.Resource{
					Name:     "3",
					Provider: "test",
				},
			},
		})
		assert.Equal(t, allocation.ResourceSlice{
			{
				Request: manifest.Resource{
					Name:     "3",
					Provider: "test",
				},
			},
		}, c)
		assert.Equal(t, allocation.ResourceSlice{
			{
				Request: manifest.Resource{
					Name:     "2",
					Provider: "test1",
				},
			},
		}, u)
		assert.Equal(t, allocation.ResourceSlice{
			{
				Request: manifest.Resource{
					Name:     "1",
					Provider: "test",
				},
			},
		}, d)
	})
}
