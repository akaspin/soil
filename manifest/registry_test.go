// +build ide test_unit

package manifest_test

import (
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestRegistry_Unmarshal(t *testing.T) {
	var pods manifest.Registry
	r, err := os.Open("testdata/example-multi.hcl")
	assert.NoError(t, err)
	defer r.Close()

	err = (&pods).Unmarshal("private", r)
	assert.NoError(t, err)
	assert.Len(t, pods, 2)
}

func TestRegistry_UnmarshalFiles(t *testing.T) {
	t.Run("0 with resources", func(t *testing.T) {
		var pods manifest.Registry
		err := pods.UnmarshalFiles("private", "testdata/test_registry_0.hcl")
		assert.NoError(t, err)
		assert.Equal(t, pods, manifest.Registry{
			&manifest.Pod{
				Namespace:  "private",
				Name:       "second",
				Runtime:    false,
				Target:     "multi-user.target",
				Constraint: map[string]string{"${meta.consul}": "true"},
				Units: []*manifest.Unit{
					{
						Transition: manifest.Transition{Create: "start", Update: "restart", Destroy: "stop", Permanent: false},
						Name:       "second-1.service",
						Source:     "[Service]\nExecStart=/usr/bin/sleep inf\n",
					},
				},
				Blobs: nil,
				Resources: []*manifest.Resource{
					{
						Name:     "8080",
						Kind:     "port",
						Required: true,
						Config:   map[string]interface{}{"fixed": "8080"},
					},
					{
						Name:     "1",
						Kind:     "counter",
						Required: true,
						Config:   map[string]interface{}{"count": "3"},
					},
					{
						Name:     "2",
						Kind:     "counter",
						Required: false,
						Config:   map[string]interface{}{"count": "1", "a": "b"},
					},
				},
			},
		})
	})
	t.Run("intro", func(t *testing.T) {
		var pods manifest.Registry
		err := pods.UnmarshalFiles("private", "testdata/files_1.hcl", "testdata/files_2.hcl")
		assert.NoError(t, err)
		assert.Len(t, pods, 3)
	})
}
