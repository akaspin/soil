// +build ide test_unit

package manifest_test

import (
	"encoding/json"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestManifest(t *testing.T) {
	var res manifest.Registry
	err := res.UnmarshalFiles("private", "testdata/example-multi.hcl")
	assert.NoError(t, err)

	t.Run("parse", func(t *testing.T) {
		assert.Equal(t, res, manifest.Registry{
			&manifest.Pod{
				Namespace: "private",
				Name:      "first",
				Runtime:   true,
				Target:    "multi-user.target",
				Units: []manifest.Unit{
					{
						Transition: manifest.Transition{Create: "start", Update: "", Destroy: "stop", Permanent: true},
						Name:       "first-1.service",
						Source:     "[Service]\n# ${meta.consul}\nExecStart=/usr/bin/sleep inf\nExecStopPost=/usr/bin/systemctl stop first-2.service\n",
					},
					{
						Transition: manifest.Transition{Create: "", Update: "start", Destroy: "", Permanent: false},
						Name:       "first-2.service",
						Source:     "[Service]\n# ${NONEXISTENT}\nExecStart=/usr/bin/sleep inf\n",
					},
				},
				Blobs: []manifest.Blob{
					{Name: "/etc/vpn/users/env", Permissions: 420, Leave: false, Source: "My file\n"},
				},
				Resources: nil,
			},
			&manifest.Pod{
				Namespace:  "private",
				Name:       "second",
				Runtime:    false,
				Target:     "multi-user.target",
				Constraint: manifest.Constraint{"${meta.consul}": "true"},
				Units: []manifest.Unit{
					{
						Transition: manifest.Transition{Create: "start", Update: "restart", Destroy: "stop", Permanent: false},
						Name:       "second-1.service",
						Source:     "[Service]\nExecStart=/usr/bin/sleep inf\n",
					},
				},
				Blobs: nil,
			},
		})

	})
	t.Run("mark", func(t *testing.T) {
		for i, mark := range []uint64{
			0xab60e0f1d3db66ec, 0xda9e24b23f46475e,
		} {
			assert.Equal(t, mark, res[i].Mark())
		}
	})
}

func TestManifest_JSON(t *testing.T) {
	var pods manifest.Registry
	err := pods.UnmarshalFiles("private", "testdata/json.hcl")
	assert.NoError(t, err)

	data, err := json.Marshal(pods[0])
	assert.Equal(t, string(data), "{\"Namespace\":\"private\",\"Name\":\"first\",\"Runtime\":true,\"Target\":\"multi-user.target\",\"Constraint\":{\"${meta.one}\":\"one\",\"${meta.two}\":\"two\"},\"Units\":[{\"Create\":\"start\",\"Update\":\"\",\"Destroy\":\"stop\",\"Permanent\":true,\"Name\":\"first-1.service\",\"Source\":\"[Service]\\n# ${meta.consul}\\nExecStart=/usr/bin/sleep inf\\nExecStopPost=/usr/bin/systemctl stop first-2.service\\n\"},{\"Create\":\"\",\"Update\":\"start\",\"Destroy\":\"\",\"Permanent\":false,\"Name\":\"first-2.service\",\"Source\":\"[Service]\\n# ${NONEXISTENT}\\nExecStart=/usr/bin/sleep inf\\n\"}],\"Blobs\":[{\"Name\":\"/etc/vpn/users/env\",\"Permissions\":420,\"Leave\":false,\"Source\":\"My file\\n\"}],\"Resources\":null}")

	// unmarshal
	pod := manifest.DefaultPod("private")
	err = json.Unmarshal(data, &pod)
	data1, err := json.Marshal(pod)
	assert.Equal(t, string(data), string(data1))
}

func TestPod_GetResource(t *testing.T) {
	var registry manifest.Registry
	err := registry.UnmarshalFiles("private", "testdata/test_pod_GetConstraint.hcl")
	assert.NoError(t, err)

	t.Run("0 request constraint", func(t *testing.T) {
		assert.Equal(t, manifest.Constraint{
			"${__resource.request.allow}":        "true",
			"${__resource.request.kind.port}":    "true",
			"${__resource.request.kind.counter}": "true",
			"${meta.consul}":                     "true",
		}, registry[0].GetResourceRequestConstraint())
		assert.Equal(t, manifest.Constraint{
			"${meta.consul}":              "true",
			"${__resource.request.allow}": "false",
		}, registry[1].GetResourceRequestConstraint())
	})
	t.Run("0 allocation constraint", func(t *testing.T) {
		assert.Equal(t, manifest.Constraint{
			"${resource.port.first.8080.allocated}": "true",
			"${resource.counter.first.1.allocated}": "true",
			"${meta.consul}":                        "true",
		}, registry[0].GetResourceAllocationConstraint())
		assert.Equal(t, manifest.Constraint{
			"${meta.consul}": "true",
		}, registry[1].GetResourceAllocationConstraint())
	})
}
