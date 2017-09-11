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

	pods, err := sd.ListPods()
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		"test-1": "/run/systemd/system/pod-cleanup-test-1.service",
		"test-2": "/run/systemd/system/pod-cleanup-test-2.service",
	}, pods)

	assert.NoError(t, sd.Cleanup())
	pods, err = sd.ListPods()
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{}, pods)
}
