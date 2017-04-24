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

	res, failures, err := manifest.ParseManifests(r)
	assert.NoError(t, err)
	assert.Len(t, failures, 0)

	t.Run("parse", func(t *testing.T) {
		assert.Equal(t, []*manifest.Pod{
			{
				Name:    "first",
				Runtime: true,
				Target:  "multi-user.target",
				Count:   2,
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
			},
			{
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
	t.Run("mark", func(t *testing.T) {
		for i, mark := range []uint64{
			0xe731e287ec137dd2, 0xb664ee7391a2f659,
		} {
			assert.Equal(t, mark, res[i].Mark())
		}
	})

}
