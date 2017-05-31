package manifest_test

import (
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestManifest(t *testing.T) {
	r, err := os.Open("testdata/example-multi.hcl")
	require.NoError(t, err)
	defer r.Close()

	res, err := manifest.ParseFromReader("private", r)
	assert.NoError(t, err)

	t.Run("parse", func(t *testing.T) {
		assert.Equal(t, []*manifest.Pod{
			{
				Namespace: "private",
				Name:    "first",
				Runtime: true,
				Target:  "multi-user.target",
				Units: []*manifest.Unit{
					{
						Transition: manifest.Transition{
							Create:  "start",
							Destroy: "stop",
						},
						Name:      "first-1.service",
						Permanent: true,
						Source:    "[Service]\n# ${meta.consul}\nExecStart=/usr/bin/sleep inf\nExecStopPost=/usr/bin/systemctl stop first-2.service\n",
					},
					{
						Transition: manifest.Transition{
							Update: "start",
						},
						Name:   "first-2.service",
						Source: "[Service]\n# ${NONEXISTENT}\nExecStart=/usr/bin/sleep inf\n",
					},
				},
				Files: []*manifest.File{
					{
						Name: "/etc/vpn/users/env",
						Permissions: 0644,
						Source: "My file\n",
					},
				},
			},
			{
				Namespace: "private",
				Name:   "second",
				Target: "default.target",
				Constraint: map[string]string{
					"${meta.consul}": "true",
				},
				Units: []*manifest.Unit{
					{
						Transition: manifest.Transition{
							Create:  "",
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
					"${meta.consul}": "true",
					"${meta.a}": "true",
				},
			},
		} {
			assert.Equal(t, map[string][]string{
				"meta": {"a", "consul"},
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
			"agent.id": "localhost",
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
			0x71bb2988cec20b6d, 0x96a775b34837d54b,
		} {
			assert.Equal(t, mark, res[i].Mark())
		}
	})

}
