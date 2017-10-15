// +build ide test_unit

package resource_test

import (
	"testing"
	"github.com/akaspin/soil/agent/resource"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
)

func TestAllocation_Clone(t *testing.T) {
	alloc := &resource.Allocation{
		PodName: "pod-1",
		Resource: &allocation.Resource{
			Request: manifest.Resource{
				Kind: "kind-1",
				Name: "res-1",
				Config: map[string]interface{}{
					"v1": 1,
				},
			},
			Values: map[string]string{
				"a": "1",
			},
		},
	}
	alloc1 := alloc.Clone()
	alloc1.Values = map[string]string{
		"a": "2",
	}
	assert.NotEqual(t, alloc1, alloc)
}
