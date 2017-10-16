// +build ide test_systemd

package fixture_test

import (
	"github.com/akaspin/soil/fixture"
	"github.com/coreos/go-systemd/dbus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewSystemd(t *testing.T) {

	sd := fixture.NewSystemd("/var/run/systemd/system", "pod-unknown")
	assert.NoError(t, sd.DeployPod("test-1", 1))
	defer sd.Cleanup()

	conn, err := dbus.New()
	assert.NoError(t, err)
	defer conn.Close()

	// try to start unit
	ch1 := make(chan string)
	_, err = conn.StartUnit("test-1-0.service", "replace", ch1)
	assert.NoError(t, err)
	res1 := <-ch1
	assert.Equal(t, "done", res1)

	// try to stop unit
	_, err = conn.StopUnit("test-1-0.service", "replace", ch1)
	assert.NoError(t, err)
	res2 := <-ch1
	assert.Equal(t, "done", res2)
}

func TestSystemd_Cleanup(t *testing.T) {
	sd := fixture.NewSystemd("/run/systemd/system", "pod-cleanup")
	assert.NoError(t, sd.DeployPod("test-1", 1))
	assert.NoError(t, sd.DeployPod("test-2", 1))

	sd.AssertUnitStates(t, []string{"pod-cleanup-*"}, map[string]string{
		"pod-cleanup-test-1.service": "active",
		"pod-cleanup-test-2.service": "active",
	})
	assert.NoError(t, sd.Cleanup())
	sd.AssertUnitStates(t, []string{"pod-cleanup-*"}, map[string]string{})
}

func TestSystemd_AssertUnitBodies(t *testing.T) {
	sd := fixture.NewSystemd("/run/systemd/system", "pod-assert-body")
	defer sd.Cleanup()
	assert.NoError(t, sd.DeployPod("test-1", 1))
	sd.AssertUnitBodies(t, []string{"pod-assert-body-*"}, map[string]string{
		"/run/systemd/system/pod-assert-body-test-1.service": `### POD test-1 {"AgentMark":456,"Namespace":"private","PodMark":123}
### UNIT /run/systemd/system/test-1-0.service {"Create":"start","Destroy":"stop","Permanent":true,"Update":"restart"}
[Unit]
Description=test-1
Before=test-1-0.service
[Service]
ExecStart=/usr/bin/sleep inf
[Install]
WantedBy=multi-user.target
`,
	})
}

func TestSystemd_AssertUnitHashes(t *testing.T) {
	sd := fixture.NewSystemd("/run/systemd/system", "pod-assert-hash")
	defer sd.Cleanup()
	assert.NoError(t, sd.DeployPod("test-1", 1))
	sd.AssertUnitHashes(t, []string{"pod-assert-hash-*"}, map[string]uint64{
		"/run/systemd/system/pod-assert-hash-test-1.service": 4712888941532877635,
	})
}
