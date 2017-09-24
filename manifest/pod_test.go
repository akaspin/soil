// +build ide test_unit

package manifest_test

import (
	"encoding/json"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestPods_Unmarshal(t *testing.T) {
	var pods manifest.Pods
	r, err := os.Open("testdata/example-multi.hcl")
	assert.NoError(t, err)
	defer r.Close()

	err = (&pods).Unmarshal("private", r)
	assert.NoError(t, err)
	assert.Len(t, pods, 2)
}

func TestManifest(t *testing.T) {
	res, err := manifest.ParseFromFiles("private", "testdata/example-multi.hcl")
	assert.NoError(t, err)

	t.Run("parse", func(t *testing.T) {
		assert.Equal(t, []*manifest.Pod{
			{
				Namespace: "private",
				Name:      "first",
				Runtime:   true,
				Target:    "multi-user.target",
				Units: []*manifest.Unit{
					{
						Transition: manifest.Transition{
							Create:    "start",
							Update:    "",
							Destroy:   "stop",
							Permanent: true,
						},
						Name:   "first-1.service",
						Source: "[Service]\n# ${meta.consul}\nExecStart=/usr/bin/sleep inf\nExecStopPost=/usr/bin/systemctl stop first-2.service\n",
					},
					{
						Transition: manifest.Transition{
							Create:  "",
							Update:  "start",
							Destroy: "",
						},
						Name:   "first-2.service",
						Source: "[Service]\n# ${NONEXISTENT}\nExecStart=/usr/bin/sleep inf\n",
					},
				},
				Blobs: []*manifest.Blob{
					{
						Name:        "/etc/vpn/users/env",
						Permissions: 0644,
						Source:      "My file\n",
					},
				},
			},
			{
				Namespace: "private",
				Name:      "second",
				Target:    "multi-user.target",
				Constraint: map[string]string{
					"${meta.consul}": "true",
				},
				Units: []*manifest.Unit{
					{
						Transition: manifest.Transition{
							Create:  "start",
							Update:  "restart",
							Destroy: "stop",
						},
						Name:   "second-1.service",
						Source: "[Service]\nExecStart=/usr/bin/sleep inf\n",
					},
				},
			},
		}, res)

	})
	t.Run("fields", func(t *testing.T) {
		for _, pod := range []*manifest.Pod{
			{
				Constraint: map[string]string{
					"${counter.test-1}": "< 4",
					"${meta.consul}":    "true",
					"${meta.a}":         "true",
				},
			},
		} {
			assert.Equal(t, map[string][]string{
				"meta":    {"a", "consul"},
				"counter": {"test-1"}},
				pod.Constraint.ExtractFields())
		}
	})
	t.Run("constraint ok", func(t *testing.T) {
		cns := manifest.Constraint(map[string]string{
			"${meta.consul}": "true",
			"${agent.id}":    "localhost",
		})
		assert.NoError(t, cns.Check(map[string]string{
			"meta.consul": "true",
			"agent.id":    "localhost",
		}))
	})
	t.Run("constraint fail", func(t *testing.T) {
		cns := manifest.Constraint(map[string]string{
			"${meta.consul}": "true",
			"${agent.id}":    "localhost",
		})
		assert.Error(t, cns.Check(map[string]string{
			"agent.id": "localhost",
		}))
	})

	t.Run("mark", func(t *testing.T) {
		for i, mark := range []uint64{
			0x1c18aee4a1c89fd0, 0x6de66aa74d55be62,
		} {
			assert.Equal(t, mark, res[i].Mark())
		}
	})
}

func TestParseFromFiles(t *testing.T) {
	pods, err := manifest.ParseFromFiles("private", "testdata/files_1.hcl", "testdata/files_2.hcl")
	assert.NoError(t, err)
	assert.Len(t, pods, 3)
}

func TestManifest_JSON(t *testing.T) {
	pods, err := manifest.ParseFromFiles("private", "testdata/json.hcl")
	assert.NoError(t, err)
	data, err := json.Marshal(pods[0])
	assert.Equal(t, "{\"Namespace\":\"private\",\"Name\":\"first\",\"Runtime\":true,\"Target\":\"multi-user.target\",\"Constraint\":{\"${meta.one}\":\"one\",\"${meta.two}\":\"two\"},\"Units\":[{\"Create\":\"start\",\"Update\":\"\",\"Destroy\":\"stop\",\"Permanent\":true,\"Name\":\"first-1.service\",\"Source\":\"[Service]\\n# ${meta.consul}\\nExecStart=/usr/bin/sleep inf\\nExecStopPost=/usr/bin/systemctl stop first-2.service\\n\"},{\"Create\":\"\",\"Update\":\"start\",\"Destroy\":\"\",\"Permanent\":false,\"Name\":\"first-2.service\",\"Source\":\"[Service]\\n# ${NONEXISTENT}\\nExecStart=/usr/bin/sleep inf\\n\"}],\"Blobs\":[{\"Name\":\"/etc/vpn/users/env\",\"Permissions\":420,\"Leave\":false,\"Source\":\"My file\\n\"}]}", string(data))

	// unmarshal
	pod := manifest.DefaultPod("private")
	err = json.Unmarshal(data, &pod)
	data1, err := json.Marshal(pod)
	assert.Equal(t, string(data), string(data1))
}
