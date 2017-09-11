// +build ide test_systemd

package allocation_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"github.com/akaspin/soil/fixture"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/manifest"
)

func TestNewFromSystemD(t *testing.T) {
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

