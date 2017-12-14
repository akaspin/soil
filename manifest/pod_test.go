// +build ide test_unit

package manifest_test

import (
	"encoding/json"
	"github.com/akaspin/soil/lib"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestManifest(t *testing.T) {
	var buffers lib.StaticBuffers
	assert.NoError(t, buffers.ReadFiles("testdata/example-multi.hcl"))
	var res manifest.Registry
	assert.NoError(t, res.Unmarshal("private", buffers.GetReaders()...))

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
			0x3d8f6a3e5d220c15, 0x4c71ae7db1bf2da7,
		} {
			assert.Equal(t, mark, res[i].Mark())
		}
	})
}

func TestManifest_JSON(t *testing.T) {
	var buffers lib.StaticBuffers
	var pods manifest.Registry
	assert.NoError(t, buffers.ReadFiles("testdata/json.hcl"))
	assert.NoError(t, pods.Unmarshal(manifest.PrivateNamespace, buffers.GetReaders()...))

	data, err := json.Marshal(pods[0])
	assert.NoError(t, err)
	assert.Equal(t, "{\"Namespace\":\"private\",\"Name\":\"first\",\"Runtime\":true,\"Target\":\"multi-user.target\",\"Constraint\":{\"${meta.one}\":\"one\",\"${meta.two}\":\"two\"},\"Units\":[{\"Create\":\"start\",\"Update\":\"\",\"Destroy\":\"stop\",\"Permanent\":true,\"Name\":\"first-1.service\",\"Source\":\"[Service]\\n# ${meta.consul}\\nExecStart=/usr/bin/sleep inf\\nExecStopPost=/usr/bin/systemctl stop first-2.service\\n\"},{\"Create\":\"\",\"Update\":\"start\",\"Destroy\":\"\",\"Permanent\":false,\"Name\":\"first-2.service\",\"Source\":\"[Service]\\n# ${NONEXISTENT}\\nExecStart=/usr/bin/sleep inf\\n\"}],\"Blobs\":[{\"Name\":\"/etc/vpn/users/env\",\"Permissions\":420,\"Leave\":false,\"Source\":\"My file\\n\"}],\"Resources\":null,\"Providers\":null}",
		string(data))

	// unmarshal
	pod := manifest.DefaultPod("private")
	err = json.Unmarshal(data, &pod)
	data1, err := json.Marshal(pod)
	assert.Equal(t, string(data), string(data1))
}
