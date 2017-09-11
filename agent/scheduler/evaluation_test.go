// +build ide test_unit

package scheduler_test

import (
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEvaluation_Plan(t *testing.T) {
	left := &allocation.Pod{
		Header: &allocation.Header{
			Name:      "pod-1",
			AgentMark: 123,
			PodMark:   456,
			Namespace: "private",
		},
		UnitFile: &allocation.UnitFile{
			Path:   "/etc/systemd/system/pod-pod-1.service",
			Source: "fake",
		},
		Units: []*allocation.Unit{
			{
				UnitFile: &allocation.UnitFile{
					Path:   "/etc/systemd/system/unit-1.service",
					Source: "fake",
				},
				Transition: &manifest.Transition{
					Create:    "start",
					Update:    "restart",
					Destroy:   "stop",
					Permanent: true,
				},
			},
			{
				UnitFile: &allocation.UnitFile{
					Path:   "/etc/systemd/system/unit-2.service",
					Source: "fake",
				},
				Transition: &manifest.Transition{
					Create:    "start",
					Update:    "restart",
					Destroy:   "stop",
					Permanent: true,
				},
			},
		},
		Blobs: []*allocation.Blob{
			{
				Name:        "/etc/test1",
				Permissions: 0644,
				Source:      "test",
			},
		},
	}

	t.Run("noop pod", func(t *testing.T) {
		right := &allocation.Pod{
			Header: &allocation.Header{
				Name:      "pod-1",
				AgentMark: 123,
				PodMark:   456,
				Namespace: "private",
			},
			UnitFile: &allocation.UnitFile{
				Path:   "/etc/systemd/system/pod-pod-1.service",
				Source: "fake",
			},
			Units: []*allocation.Unit{
				{
					UnitFile: &allocation.UnitFile{
						Path:   "/etc/systemd/system/unit-1.service",
						Source: "fake",
					},
					Transition: &manifest.Transition{
						Permanent: true,
						Create:    "start",
						Update:    "restart",
						Destroy:   "stop",
					},
				},
				{
					UnitFile: &allocation.UnitFile{
						Path:   "/etc/systemd/system/unit-2.service",
						Source: "fake",
					},
					Transition: &manifest.Transition{
						Permanent: true,
						Create:    "start",
						Update:    "restart",
						Destroy:   "stop",
					},
				},
			},
			Blobs: []*allocation.Blob{
				{
					Name:        "/etc/test1",
					Permissions: 0644,
					Source:      "test",
				},
			},
		}
		evaluation := scheduler.NewEvaluation(left, right)
		assert.Equal(t, "pod-1:[]", evaluation.String())
	})
	t.Run("unit-1 perm to disabled", func(t *testing.T) {
		right := &allocation.Pod{
			Header: &allocation.Header{
				Name:      "pod-1",
				AgentMark: 123,
				PodMark:   456,
				Namespace: "private",
			},
			UnitFile: &allocation.UnitFile{
				Path:   "/etc/systemd/system/pod-pod-1.service",
				Source: "fake",
			},
			Units: []*allocation.Unit{
				{
					UnitFile: &allocation.UnitFile{
						Path:   "/etc/systemd/system/unit-1.service",
						Source: "fake",
					},
					Transition: &manifest.Transition{
						Permanent: false,
						Create:    "start",
						Update:    "restart",
						Destroy:   "stop",
					},
				},
				{
					UnitFile: &allocation.UnitFile{
						Path:   "/etc/systemd/system/unit-2.service",
						Source: "fake",
					},
					Transition: &manifest.Transition{
						Permanent: true,
						Create:    "start",
						Update:    "restart",
						Destroy:   "stop",
					},
				},
			},
		}
		evaluation := scheduler.NewEvaluation(left, right)
		assert.Equal(t, "pod-1:[3:disable-unit:/etc/systemd/system/unit-1.service 5:delete-blob:/etc/test1]", evaluation.String())
	})
	t.Run("update unit-1 and file", func(t *testing.T) {
		right := &allocation.Pod{
			Header: &allocation.Header{
				Name:      "pod-1",
				AgentMark: 123,
				PodMark:   456,
				Namespace: "private",
			},
			UnitFile: &allocation.UnitFile{
				Path:   "/etc/systemd/system/pod-pod-1.service",
				Source: "fake",
			},
			Units: []*allocation.Unit{
				{
					UnitFile: &allocation.UnitFile{
						Path:   "/etc/systemd/system/unit-1.service",
						Source: "fake1",
					},
					Transition: &manifest.Transition{
						Permanent: true,
						Create:    "start",
						Update:    "restart",
						Destroy:   "stop",
					},
				},
				{
					UnitFile: &allocation.UnitFile{
						Path:   "/etc/systemd/system/unit-2.service",
						Source: "fake",
					},
					Transition: &manifest.Transition{
						Permanent: true,
						Create:    "start",
						Update:    "restart",
						Destroy:   "stop",
					},
				},
			},
			Blobs: []*allocation.Blob{
				{
					Name:        "/etc/test1",
					Permissions: 0644,
					Source:      "test1",
				},
			},
		}
		evaluation := scheduler.NewEvaluation(left, right)
		assert.Equal(t, "pod-1:[2:write-blob:/etc/test1 2:write-unit:/etc/systemd/system/unit-1.service 3:enable-unit:/etc/systemd/system/unit-1.service 4:restart:/etc/systemd/system/unit-1.service]", evaluation.String())
	})
	t.Run("create pod form left", func(t *testing.T) {
		evaluation := scheduler.NewEvaluation(nil, left)
		assert.Equal(t, "pod-1:[2:write-blob:/etc/test1 2:write-unit:/etc/systemd/system/pod-pod-1.service 2:write-unit:/etc/systemd/system/unit-1.service 2:write-unit:/etc/systemd/system/unit-2.service 3:enable-unit:/etc/systemd/system/pod-pod-1.service 3:enable-unit:/etc/systemd/system/unit-1.service 3:enable-unit:/etc/systemd/system/unit-2.service 4:start:/etc/systemd/system/pod-pod-1.service 4:start:/etc/systemd/system/unit-1.service 4:start:/etc/systemd/system/unit-2.service]", evaluation.String())
	})
	t.Run("destroy pod", func(t *testing.T) {
		evaluation := scheduler.NewEvaluation(left, nil)
		assert.Equal(t, "pod-1:[0:stop:/etc/systemd/system/pod-pod-1.service 0:stop:/etc/systemd/system/unit-1.service 0:stop:/etc/systemd/system/unit-2.service 1:delete-unit:/etc/systemd/system/pod-pod-1.service 1:delete-unit:/etc/systemd/system/unit-1.service 1:delete-unit:/etc/systemd/system/unit-2.service 5:delete-blob:/etc/test1]", evaluation.String())
	})
	t.Run("change prefix", func(t *testing.T) {
		right := &allocation.Pod{
			Header: &allocation.Header{
				Name:      "pod-1",
				AgentMark: 123,
				PodMark:   456,
				Namespace: "private",
			},
			UnitFile: &allocation.UnitFile{
				Path:   "/etc/systemd/system/pod-local-pod-1.service",
				Source: "fake",
			},
			Units: []*allocation.Unit{
				{
					UnitFile: &allocation.UnitFile{
						Path:   "/etc/systemd/system/unit-1.service",
						Source: "fake",
					},
					Transition: &manifest.Transition{
						Permanent: true,
						Create:    "start",
						Update:    "restart",
						Destroy:   "stop",
					},
				},
				{
					UnitFile: &allocation.UnitFile{
						Path:   "/etc/systemd/system/unit-2.service",
						Source: "fake",
					},
					Transition: &manifest.Transition{
						Permanent: true,
						Create:    "start",
						Update:    "restart",
						Destroy:   "stop",
					},
				},
			},
			Blobs: []*allocation.Blob{
				{
					Name:        "/etc/test1",
					Permissions: 0644,
					Source:      "test",
				},
			},
		}
		evaluation := scheduler.NewEvaluation(left, right)
		assert.Equal(t, "pod-1:[0:stop:/etc/systemd/system/pod-pod-1.service 1:delete-unit:/etc/systemd/system/pod-pod-1.service 2:write-unit:/etc/systemd/system/pod-local-pod-1.service 3:enable-unit:/etc/systemd/system/pod-local-pod-1.service 4:start:/etc/systemd/system/pod-local-pod-1.service]", evaluation.String())
	})
	t.Run("local to runtime", func(t *testing.T) {
		right := &allocation.Pod{
			Header: &allocation.Header{
				Name:      "pod-1",
				AgentMark: 123,
				PodMark:   456,
				Namespace: "private",
			},
			UnitFile: &allocation.UnitFile{
				Path:   "/run/systemd/system/pod-pod-1.service",
				Source: "fake",
			},
			Units: []*allocation.Unit{
				{
					UnitFile: &allocation.UnitFile{
						Path:   "/run/systemd/system/unit-1.service",
						Source: "fake",
					},
					Transition: &manifest.Transition{
						Create:    "start",
						Update:    "restart",
						Destroy:   "stop",
						Permanent: true,
					},
				},
				{
					UnitFile: &allocation.UnitFile{
						Path:   "/run/systemd/system/unit-2.service",
						Source: "fake",
					},
					Transition: &manifest.Transition{
						Create:    "start",
						Update:    "restart",
						Destroy:   "stop",
						Permanent: true,
					},
				},
			},
			Blobs: []*allocation.Blob{
				{
					Name:        "/etc/test1",
					Permissions: 0644,
					Source:      "test",
				},
			},
		}
		evaluation := scheduler.NewEvaluation(left, right)
		assert.Equal(t, "pod-1:[0:stop:/etc/systemd/system/pod-pod-1.service 0:stop:/etc/systemd/system/unit-1.service 0:stop:/etc/systemd/system/unit-2.service 1:delete-unit:/etc/systemd/system/pod-pod-1.service 1:delete-unit:/etc/systemd/system/unit-1.service 1:delete-unit:/etc/systemd/system/unit-2.service 2:write-unit:/run/systemd/system/pod-pod-1.service 2:write-unit:/run/systemd/system/unit-1.service 2:write-unit:/run/systemd/system/unit-2.service 3:enable-unit:/run/systemd/system/pod-pod-1.service 3:enable-unit:/run/systemd/system/unit-1.service 3:enable-unit:/run/systemd/system/unit-2.service 4:start:/run/systemd/system/pod-pod-1.service 4:start:/run/systemd/system/unit-1.service 4:start:/run/systemd/system/unit-2.service]", evaluation.String())
	})
}
