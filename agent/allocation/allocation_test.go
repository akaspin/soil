// +build ide test_unit

package allocation_test

import (
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewFromManifest(t *testing.T) {

	env := map[string]string{
		"meta.consul":     "true",
		"system.pod_exec": "ExecStart=/usr/bin/sleep inf",
	}

	t.Run("0 simple names", func(t *testing.T) {
		var pods manifest.Registry
		err := pods.UnmarshalFiles("private", "testdata/test_new_from_manifest_0.hcl")
		assert.NoError(t, err)

		m := pods[0]
		var res *allocation.Pod
		res, err = allocation.NewFromManifest(m, env)
		assert.NoError(t, err)
		assert.Equal(t, &allocation.Pod{
			Header: &allocation.Header{
				Name:      "pod-1",
				PodMark:   7228519356168739269,
				AgentMark: 7076960218577909541,
				Namespace: "private",
			},
			UnitFile: &allocation.UnitFile{
				Path:   "/run/systemd/system/pod-private-pod-1.service",
				Source: "### POD pod-1 {\"AgentMark\":7076960218577909541,\"Namespace\":\"private\",\"PodMark\":7228519356168739269}\n### UNIT /run/systemd/system/unit-1.service {\"Create\":\"start\",\"Update\":\"\",\"Destroy\":\"stop\",\"Permanent\":false}\n### UNIT /run/systemd/system/unit-2.service {\"Create\":\"start\",\"Update\":\"\",\"Destroy\":\"stop\",\"Permanent\":false}\n### BLOB /etc/test {\"Leave\":false,\"Permissions\":420}\n\n[Unit]\nDescription=pod-1\nBefore=unit-1.service unit-2.service\n[Service]\nExecStart=/usr/bin/sleep inf\n[Install]\nWantedBy=multi-user.target\n",
			},
			Units: []*allocation.Unit{
				{
					Transition: &manifest.Transition{
						Create:    "start",
						Destroy:   "stop",
						Permanent: false,
					},
					UnitFile: &allocation.UnitFile{
						Path:   "/run/systemd/system/unit-1.service",
						Source: "# true",
					},
				},
				{
					Transition: &manifest.Transition{
						Create:    "start",
						Destroy:   "stop",
						Permanent: false,
					},
					UnitFile: &allocation.UnitFile{
						Path:   "/run/systemd/system/unit-2.service",
						Source: "# true 10090666253179731817",
					},
				},
			},
			Blobs: []*allocation.Blob{
				{
					Name:        "/etc/test",
					Permissions: 0644,
					Source:      "test",
				},
			},
		}, res)
	})
	t.Run("interpolate names", func(t *testing.T) {
		var pods manifest.Registry
		err := pods.UnmarshalFiles("private", "testdata/test_new_from_manifest_1.hcl")
		assert.NoError(t, err)
		m := pods[0]
		var res *allocation.Pod
		res, err = allocation.NewFromManifest(m, env)
		assert.NoError(t, err)

		assert.Equal(t, res, &allocation.Pod{
			Header:   &allocation.Header{Name: "pod-2", PodMark: 0xd5f515f9af5de917, AgentMark: 0x623669d2cde83725, Namespace: "private"},
			UnitFile: &allocation.UnitFile{Path: "/run/systemd/system/pod-private-pod-2.service", Source: "### POD pod-2 {\"AgentMark\":7076960218577909541,\"Namespace\":\"private\",\"PodMark\":15417253061505968407}\n### UNIT /run/systemd/system/pod-2-unit-1.service {\"Create\":\"start\",\"Update\":\"\",\"Destroy\":\"stop\",\"Permanent\":false}\n### UNIT /run/systemd/system/private-unit-2.service {\"Create\":\"start\",\"Update\":\"\",\"Destroy\":\"stop\",\"Permanent\":false}\n### BLOB /pod-2/etc/test {\"Leave\":false,\"Permissions\":420}\n\n[Unit]\nDescription=pod-2\nBefore=pod-2-unit-1.service private-unit-2.service\n[Service]\nExecStart=/usr/bin/sleep inf\n[Install]\nWantedBy=multi-user.target\n"},
			Units: []*allocation.Unit{
				&allocation.Unit{
					UnitFile:   &allocation.UnitFile{Path: "/run/systemd/system/pod-2-unit-1.service", Source: "# true multi-user.target"},
					Transition: &manifest.Transition{Create: "start", Update: "", Destroy: "stop", Permanent: false},
				},
				&allocation.Unit{
					UnitFile:   &allocation.UnitFile{Path: "/run/systemd/system/private-unit-2.service", Source: "# true 10090666253179731817"},
					Transition: &manifest.Transition{Create: "start", Update: "", Destroy: "stop", Permanent: false},
				},
			},
			Blobs: []*allocation.Blob{
				&allocation.Blob{Name: "/pod-2/etc/test", Permissions: 420, Leave: false, Source: "test"},
			},
		})
	})
}

func TestHeader_Unmarshal(t *testing.T) {
	src := `### POD pod-1 {"AgentMark":123,"Namespace":"private","PodMark":345}
### UNIT /etc/systemd/system/unit-1.service {"Create":"start","Update":"","Destroy":"","Permanent":true}
### UNIT /etc/systemd/system/unit-2.service {"Create":"","Update":"start","Destroy":"","Permanent":false}
### BLOB /etc/test {"Leave":true,"Permissions":420}
[Unit]
`
	header := &allocation.Header{}
	units, blobs, err := header.Unmarshal(src)
	assert.NoError(t, err)
	assert.Equal(t, []*allocation.Unit{
		{
			UnitFile: &allocation.UnitFile{
				Path: "/etc/systemd/system/unit-1.service",
			},
			Transition: &manifest.Transition{
				Create:    "start",
				Permanent: true,
			},
		},
		{
			UnitFile: &allocation.UnitFile{
				Path: "/etc/systemd/system/unit-2.service",
			},
			Transition: &manifest.Transition{
				Update:    "start",
				Permanent: false,
			},
		},
	}, units)
	assert.Equal(t, []*allocation.Blob{
		{
			Name:        "/etc/test",
			Leave:       true,
			Permissions: 0644,
		},
	}, blobs)
	assert.Equal(t, &allocation.Header{
		Name:      "pod-1",
		AgentMark: 123,
		PodMark:   345,
		Namespace: "private",
	}, header)
}

func TestHeader_Marshal(t *testing.T) {
	units := []*allocation.Unit{
		{
			Transition: &manifest.Transition{
				Create:    "start",
				Permanent: true,
			},
			UnitFile: &allocation.UnitFile{
				Path: "/etc/systemd/system/unit-1.service",
			},
		},
	}
	blobs := []*allocation.Blob{
		{
			Name:        "/etc/test",
			Permissions: 0644,
			Source:      "my-file",
		},
	}
	h := &allocation.Header{
		Namespace: "private",
		AgentMark: 234,
		PodMark:   123,
	}
	res, err := h.Marshal("pod-1", units, blobs)
	assert.NoError(t, err)
	assert.Equal(t, "### POD pod-1 {\"AgentMark\":234,\"Namespace\":\"private\",\"PodMark\":123}\n### UNIT /etc/systemd/system/unit-1.service {\"Create\":\"start\",\"Update\":\"\",\"Destroy\":\"\",\"Permanent\":true}\n### BLOB /etc/test {\"Leave\":false,\"Permissions\":420}\n", res)
}
