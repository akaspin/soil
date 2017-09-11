package allocation_test

import (
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/fixture"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewFromSystemD(t *testing.T) {
	fixture.RunTestUnless(t, "TEST_INTEGRATION")
	fixture.RunTestIf(t, "TEST_SYSTEMD")

	sd := fixture.NewSystemd("/run/systemd/system", "pod")
	defer sd.Cleanup()
	assert.NoError(t, sd.DeployPod("test-1", 2))

	alloc, err := allocation.NewFromSystemD("/run/systemd/system/pod-test-1.service")
	assert.NoError(t, err)
	assert.Equal(t, &allocation.Pod{
		Header: &allocation.Header{
			Name:      "test-1",
			PodMark:   123,
			AgentMark: 456,
			Namespace: "private",
		},
		UnitFile: &allocation.UnitFile{
			Path: "/run/systemd/system/pod-test-1.service",
			Source: `### POD test-1 {"AgentMark":456,"Namespace":"private","PodMark":123}
### UNIT /run/systemd/system/test-1-0.service {"Create":"start","Destroy":"stop","Permanent":true,"Update":"restart"}
### UNIT /run/systemd/system/test-1-1.service {"Create":"start","Destroy":"stop","Permanent":true,"Update":"restart"}
[Unit]
Description=test-1
Before=test-1-0.service test-1-1.service
[Service]
ExecStart=/usr/bin/sleep inf
[Install]
WantedBy=multi-user.target
`,
		},
		Units: []*allocation.Unit{
			{
				UnitFile: &allocation.UnitFile{
					Path: "/run/systemd/system/test-1-0.service",
					Source: `[Unit]
Description=Unit test-1-0.service
[Service]
ExecStart=/usr/bin/sleep inf
[Install]
WantedBy=multi-user.target
`,
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
					Path: "/run/systemd/system/test-1-1.service",
					Source: `[Unit]
Description=Unit test-1-1.service
[Service]
ExecStart=/usr/bin/sleep inf
[Install]
WantedBy=multi-user.target
`,
				},
				Transition: &manifest.Transition{
					Create:    "start",
					Update:    "restart",
					Destroy:   "stop",
					Permanent: true,
				},
			},
		},
	}, alloc)
}

func TestNewFromManifest(t *testing.T) {
	m := &manifest.Pod{
		Runtime:   true,
		Namespace: "private",
		Name:      "pod-1",
		Target:    "multi-user.target",
		Units: []*manifest.Unit{
			{
				Name:   "unit-1.service",
				Source: `# ${meta.consul}`,
				Transition: manifest.Transition{
					Create:  "start",
					Destroy: "stop",
				},
			},
			{
				Name:   "unit-2.service",
				Source: `# ${meta.consul} ${blob.etc-test}`,
				Transition: manifest.Transition{
					Create:  "start",
					Destroy: "stop",
				},
			},
		},
		Blobs: []*manifest.Blob{
			{
				Name:        "/etc/test",
				Permissions: 0644,
				Source:      "test",
			},
		},
	}
	env := map[string]string{
		"meta.consul":    "true",
		"agent.pod_exec": "ExecStart=/usr/bin/sleep inf",
	}
	res, err := allocation.NewFromManifest(m, env)
	assert.NoError(t, err)
	assert.Equal(t, &allocation.Pod{
		Header: &allocation.Header{
			Name:      "pod-1",
			PodMark:   7228519356168739269,
			AgentMark: 13519672434109364665,
			Namespace: "private",
		},
		UnitFile: &allocation.UnitFile{
			Path:   "/run/systemd/system/pod-private-pod-1.service",
			Source: "### POD pod-1 {\"AgentMark\":13519672434109364665,\"Namespace\":\"private\",\"PodMark\":7228519356168739269}\n### UNIT /run/systemd/system/unit-1.service {\"Create\":\"start\",\"Update\":\"\",\"Destroy\":\"stop\",\"Permanent\":false}\n### UNIT /run/systemd/system/unit-2.service {\"Create\":\"start\",\"Update\":\"\",\"Destroy\":\"stop\",\"Permanent\":false}\n### BLOB /etc/test {\"Leave\":false,\"Permissions\":420}\n\n[Unit]\nDescription=pod-1\nBefore=unit-1.service unit-2.service\n[Service]\nExecStart=/usr/bin/sleep inf\n[Install]\nWantedBy=multi-user.target\n",
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
