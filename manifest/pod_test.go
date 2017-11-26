// +build ide test_unit

package manifest_test

import (
	"encoding/json"
	"github.com/akaspin/soil/lib"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPods_Unmarshal(t *testing.T) {
	t.Run(`0 complex`, func(t *testing.T) {
		var buffers lib.StaticBuffers
		assert.Error(t, buffers.ReadFiles(
			"testdata/TestPods_Unmarshal_0_0.hcl",
			"testdata/nonexistent.hcl",
			"testdata/TestPods_Unmarshal_0_1.hcl",
			"testdata/TestPods_Unmarshal_0_2.hcl",
		))
		var res manifest.Pods
		assert.Error(t, res.Unmarshal("private", buffers.GetReaders()...))
		assert.Equal(t, manifest.Pods{
			{
				Namespace: manifest.PrivateNamespace,
				Name:      "1",
				Runtime:   true,
				Target:    "multi-user.target",
				Units: manifest.Units{
					{
						Name: "1",
						Transition: manifest.Transition{
							Create: "start",
							Update: "restart", Destroy: "stop", Permanent: false},
					},
					{
						Name: "2",
						Transition: manifest.Transition{
							Create: "start",
							Update: "restart", Destroy: "stop", Permanent: false},
					},
				},
			},
			{
				Namespace: manifest.PrivateNamespace,
				Name:      "2",
				Runtime:   true,
				Target:    "multi-user.target",
			},
		}, res)
	})
}

func TestManifest(t *testing.T) {

	t.Run("parse", func(t *testing.T) {
		var buffers lib.StaticBuffers
		assert.NoError(t, buffers.ReadFiles("testdata/example-multi.hcl"))
		var res manifest.Pods
		assert.NoError(t, res.Unmarshal("private", buffers.GetReaders()...))
		assert.Equal(t, res, manifest.Pods{
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
		var buffers lib.StaticBuffers
		assert.NoError(t, buffers.ReadFiles("testdata/example-multi.hcl"))
		var res manifest.Pods
		assert.NoError(t, res.Unmarshal("private", buffers.GetReaders()...))
		for i, mark := range []uint64{
			0x3d8f6a3e5d220c15, 0x4c71ae7db1bf2da7,
		} {
			assert.Equal(t, mark, res[i].Mark())
		}
	})
	t.Run("0 with resources", func(t *testing.T) {
		var buffers lib.StaticBuffers
		var pods manifest.Pods
		assert.NoError(t, buffers.ReadFiles("testdata/test_registry_0.hcl"))
		assert.NoError(t, pods.Unmarshal(manifest.PrivateNamespace, buffers.GetReaders()...))
		assert.Equal(t, pods, manifest.Pods{
			&manifest.Pod{
				Namespace:  "private",
				Name:       "second",
				Runtime:    false,
				Target:     "multi-user.target",
				Constraint: map[string]string{"${meta.consul}": "true"},
				Units: []manifest.Unit{
					{
						Transition: manifest.Transition{Create: "start", Update: "restart", Destroy: "stop", Permanent: false},
						Name:       "second-1.service",
						Source:     "[Service]\nExecStart=/usr/bin/sleep inf\n",
					},
				},
				Blobs: nil,
				Resources: []manifest.Resource{
					{
						Name:   "1",
						Kind:   "counter",
						Config: map[string]interface{}{"count": "3"},
					},
					{
						Name:   "2",
						Kind:   "counter",
						Config: map[string]interface{}{"count": "1", "a": "b"},
					},
					{
						Name:   "8080",
						Kind:   "port",
						Config: map[string]interface{}{"fixed": "8080"},
					},
				},
			},
		})
	})
	t.Run("intro", func(t *testing.T) {
		var buffers lib.StaticBuffers
		var pods manifest.Pods
		assert.NoError(t, buffers.ReadFiles("testdata/files_1.hcl", "testdata/files_2.hcl"))
		assert.NoError(t, pods.Unmarshal(manifest.PrivateNamespace, buffers.GetReaders()...))
		assert.Len(t, pods, 3)
	})
}

func TestManifest_JSON(t *testing.T) {
	var buffers lib.StaticBuffers
	var pods manifest.Pods
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
