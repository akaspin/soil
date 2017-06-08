package manifest_test

import (
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestCompare(t *testing.T) {
	left := "2"
	right := "10"
	assert.False(t, left < right)

	left1, err := strconv.ParseFloat(left, 64)
	assert.NoError(t, err)
	right1, err := strconv.ParseFloat(right, 64)
	assert.NoError(t, err)
	assert.True(t, left1 < right1)
}

func TestSplit(t *testing.T) {
	assert.Len(t, strings.SplitN("1000", " ", 2), 1)
	assert.Len(t, strings.SplitN("< 1000", " ", 2), 2)
}

func TestConstraint_Check(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		constraint := manifest.Constraint{
			"one,two": "${meta.field}",
		}
		t.Log()
		assert.NoError(t, constraint.Check(map[string]string{
			"meta.field": "one,two",
		}))
	})
	t.Run("in ok", func(t *testing.T) {
		constraint := manifest.Constraint{
			"one,two": "~ ${meta.field}",
		}
		assert.NoError(t, constraint.Check(map[string]string{
			"meta.field": "one,two,three",
		}))
	})
	t.Run("in fail", func(t *testing.T) {
		constraint := manifest.Constraint{
			"one,two": "~ ${meta.field}",
		}
		assert.Error(t, constraint.Check(map[string]string{
			"meta.field": "one,three",
		}))
	})
	t.Run("less ok", func(t *testing.T) {
		constraint := manifest.Constraint{
			"2": "< ${meta.num}",
		}
		assert.NoError(t, constraint.Check(map[string]string{
			"meta.num": "11",
		}))
	})
	t.Run("less fail", func(t *testing.T) {
		constraint := manifest.Constraint{
			"2": "< ${meta.num}",
		}
		assert.Error(t, constraint.Check(map[string]string{
			"meta.num": "1",
		}))
	})
	t.Run("greater ok", func(t *testing.T) {
		constraint := manifest.Constraint{
			"2": "> ${meta.num}",
		}
		assert.NoError(t, constraint.Check(map[string]string{
			"meta.num": "1",
		}))
	})
	t.Run("greater fail", func(t *testing.T) {
		constraint := manifest.Constraint{
			"2": "> ${meta.num}",
		}
		assert.Error(t, constraint.Check(map[string]string{
			"meta.num": "3",
		}))
	})
	t.Run("empty", func(t *testing.T) {
		constraint := manifest.Constraint{}
		assert.NoError(t, constraint.Check(map[string]string{
			"meta.num": "3",
		}))
	})

	return
}

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
				Name:      "first",
				Runtime:   true,
				Target:    "multi-user.target",
				Units: []*manifest.Unit{
					{
						Transition: manifest.Transition{
							Create:  "start",
							Update:  "",
							Destroy: "stop",
						},
						Name:      "first-1.service",
						Permanent: true,
						Source:    "[Service]\n# ${meta.consul}\nExecStart=/usr/bin/sleep inf\nExecStopPost=/usr/bin/systemctl stop first-2.service\n",
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
			0x929c7bc2b806e194, 0xa28a0e338d4eb333,
		} {
			assert.Equal(t, mark, res[i].Mark())
		}
	})

}
