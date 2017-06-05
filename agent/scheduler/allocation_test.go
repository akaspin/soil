package scheduler_test

import (
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/manifest"
	"github.com/mitchellh/hashstructure"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewFromManifest(t *testing.T) {
	m := &manifest.Pod{
		Runtime: true,
		Namespace: "private",
		Name: "pod-1",
		Target: "multi-user.target",
		Units: []*manifest.Unit{
			{
				Name: "unit-1.service",
				Source: `# ${meta.consul}`,
				Transition: manifest.Transition{
					Create: "start",
					Destroy: "stop",
				},
			},
			{
				Name: "unit-2.service",
				Source: `# ${meta.consul} ${blob.etc-test}`,
				Transition: manifest.Transition{
					Create: "start",
					Destroy: "stop",
				},
			},
		},
		Blobs: []*manifest.Blob{
			{
				Name: "/etc/test",
				Permissions: 0644,
				Source: "test",
			},
		},
	}
	env := map[string]string{
		"meta.consul": "true",
		"agent.pod_exec": "ExecStart=/usr/bin/sleep inf",
	}
	mark, _ := hashstructure.Hash(env, nil)

	res, err := scheduler.NewAllocationFromManifest(m, env, mark)
	assert.NoError(t, err)
	assert.Equal(t, &scheduler.Allocation{
		AllocationHeader: &scheduler.AllocationHeader{
			Name: "pod-1",
			PodMark: 8958585432400940686,
			AgentMark: 13519672434109364665,
			Namespace: "private",
		},
		AllocationFile: &scheduler.AllocationFile{
			Path: "/run/systemd/system/pod-private-pod-1.service",
			Source: "### POD pod-1 {\"AgentMark\":13519672434109364665,\"Namespace\":\"private\",\"PodMark\":8958585432400940686}\n### UNIT /run/systemd/system/unit-1.service {\"Create\":\"start\",\"Update\":\"\",\"Destroy\":\"stop\",\"Permanent\":false}\n### UNIT /run/systemd/system/unit-2.service {\"Create\":\"start\",\"Update\":\"\",\"Destroy\":\"stop\",\"Permanent\":false}\n### BLOB /etc/test {\"Leave\":false,\"Permissions\":420}\n\n[Unit]\nDescription=pod-1\nBefore=unit-1.service unit-2.service\n[Service]\nExecStart=/usr/bin/sleep inf\n[Install]\nWantedBy=multi-user.target\n",
		},
		Units: []*scheduler.AllocationUnit{
			{
				AllocationUnitHeader: &scheduler.AllocationUnitHeader{
					Permanent: false,
					Transition: manifest.Transition{
						Create: "start",
						Destroy: "stop",
					},
				},
				AllocationFile: &scheduler.AllocationFile{
					Path: "/run/systemd/system/unit-1.service",
					Source: "# true",
				},
			},
			{
				AllocationUnitHeader: &scheduler.AllocationUnitHeader{
					Permanent: false,
					Transition: manifest.Transition{
						Create: "start",
						Destroy: "stop",
					},
				},
				AllocationFile: &scheduler.AllocationFile{
					Path: "/run/systemd/system/unit-2.service",
					Source: "# true 10090666253179731817",
				},
			},
		},
		Blobs: []*scheduler.AllocationBlob{
			{
				Name: "/etc/test",
				Permissions: 0644,
				Source: "test",
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
	header := &scheduler.AllocationHeader{}
	units, blobs, err := header.Unmarshal(src)
	assert.NoError(t, err)
	assert.Equal(t, []*scheduler.AllocationUnit{
		{
			AllocationFile: &scheduler.AllocationFile{
				Path: "/etc/systemd/system/unit-1.service",
			},
			AllocationUnitHeader: &scheduler.AllocationUnitHeader{
				Permanent: true,
				Transition: manifest.Transition{
					Create: "start",
				},
			},
		},
		{
			AllocationFile: &scheduler.AllocationFile{
				Path: "/etc/systemd/system/unit-2.service",
			},
			AllocationUnitHeader: &scheduler.AllocationUnitHeader{
				Permanent: false,
				Transition: manifest.Transition{
					Update: "start",
				},
			},
		},
	}, units)
	assert.Equal(t, []*scheduler.AllocationBlob{
		{
			Name: "/etc/test",
			Leave: true,
			Permissions: 0644,
		},
	}, blobs)
	assert.Equal(t, &scheduler.AllocationHeader{
		Name:      "pod-1",
		AgentMark: 123,
		PodMark:   345,
		Namespace: "private",
	}, header)
}

func TestHeader_Marshal(t *testing.T) {
	units := []*scheduler.AllocationUnit{
		{
			AllocationUnitHeader: &scheduler.AllocationUnitHeader{
				Permanent: true,
				Transition: manifest.Transition{
					Create: "start",
				},
			},
			AllocationFile: &scheduler.AllocationFile{
				Path: "/etc/systemd/system/unit-1.service",
			},
		},
	}
	blobs := []*scheduler.AllocationBlob{
		{
			Name: "/etc/test",
			Permissions: 0644,
			Source: "my-file",
		},
	}
	h := &scheduler.AllocationHeader{
		Namespace: "private",
		AgentMark: 234,
		PodMark:   123,
	}
	res, err := h.Marshal("pod-1", units, blobs)
	assert.NoError(t, err)
	assert.Equal(t, "### POD pod-1 {\"AgentMark\":234,\"Namespace\":\"private\",\"PodMark\":123}\n### UNIT /etc/systemd/system/unit-1.service {\"Create\":\"start\",\"Update\":\"\",\"Destroy\":\"\",\"Permanent\":true}\n### BLOB /etc/test {\"Leave\":false,\"Permissions\":420}\n", res)
}


