// +build ide test_systemd

package scheduler_test

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/fixture"
	"github.com/akaspin/soil/manifest"
	"github.com/coreos/go-systemd/dbus"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func assertUnits(names []string, states map[string]string) (err error) {
	conn, err := dbus.New()
	if err != nil {
		return
	}
	defer conn.Close()
	l, err := conn.ListUnitsByPatterns([]string{}, names)
	if err != nil {
		return
	}
	res := map[string]string{}
	for _, u := range l {
		res[u.Name] = u.ActiveState
	}
	if !reflect.DeepEqual(states, res) {
		err = fmt.Errorf("not equal %v %v", states, res)
	}
	return
}

func TestNewEvaluator(t *testing.T) {
	sd := fixture.NewSystemd("/run/systemd/system", "pod")
	defer sd.Cleanup()
	assert.NoError(t, sd.DeployPod("test-1", 3))
	assert.NoError(t, sd.DeployPod("test-2", 3))

	ctx := context.Background()
	evaluator := scheduler.NewEvaluator(ctx, logx.GetLog("test"))

	assert.NoError(t, evaluator.Open())
	time.Sleep(time.Second)

	assert.Equal(t, map[string]*allocation.Header{
		"test-1": {
			Name:      "test-1",
			Namespace: "private",
			PodMark:   123,
			AgentMark: 456,
		},
		"test-2": {
			Name:      "test-2",
			Namespace: "private",
			PodMark:   123,
			AgentMark: 456,
		},
	}, evaluator.List())

	assert.NoError(t, evaluator.Close())
	assert.NoError(t, evaluator.Wait())
}

func TestEvaluator_Submit(t *testing.T) {
	conn, err := dbus.New()
	assert.NoError(t, err)
	defer conn.Close()

	sd := fixture.NewSystemd("/run/systemd/system", "pod")
	sd.Cleanup()
	defer sd.Cleanup()

	ctx := context.Background()
	evaluator := scheduler.NewEvaluator(ctx, logx.GetLog("test"))
	assert.NoError(t, evaluator.Open())

	time.Sleep(time.Second)

	t.Run("create pod-1", func(t *testing.T) {
		alloc := &allocation.Pod{
			Header: &allocation.Header{
				Name:      "pod-1",
				PodMark:   1,
				AgentMark: 0,
				Namespace: "private",
			},
			UnitFile: &allocation.UnitFile{
				SystemPaths: allocation.DefaultSystemDPaths(),
				Path:        "/run/systemd/system/pod-private-pod-1.service",
				Source: `### POD pod-1 {"AgentMark":0,"PodMark":1,"Namespace":"private"}
### UNIT unit-1.service {"Permanent":false,"Create":"start","Update":"restart","Destroy":"stop"}
[Unit]
Description=Pod
Before=unit-1.service
[Service]
ExecStart=/usr/bin/sleep inf
[Install]
WantedBy=default.target
`,
			},
			Units: []*allocation.Unit{
				{
					UnitFile: &allocation.UnitFile{
						SystemPaths: allocation.DefaultSystemDPaths(),
						Path:        "/run/systemd/system/unit-1.service",
						Source: `
[Unit]
Description=Unit 1
[Service]
ExecStart=/usr/bin/sleep inf
[Install]
WantedBy=default.target
`,
					},
					Transition: &manifest.Transition{
						Permanent: false,
						Create:    "start",
						Update:    "restart",
						Destroy:   "stop",
					},
				},
			},
		}
		evaluator.Submit("pod-1", alloc)
		time.Sleep(time.Second)

		assert.NoError(t, assertUnits(
			[]string{"pod-private-pod-1.service", "unit-1.service"},
			map[string]string{
				"pod-private-pod-1.service": "active",
				"unit-1.service":            "active",
			}))
		assert.Equal(t, map[string]*allocation.Header{
			"pod-1": {
				Name:      "pod-1",
				Namespace: "private",
				PodMark:   1,
				AgentMark: 0,
			},
		}, evaluator.List())
	})
	t.Run("destroy non-existent", func(t *testing.T) {
		evaluator.Submit("pod-2", nil)
		time.Sleep(time.Second)
		assert.NoError(t, assertUnits(
			[]string{"pod-private-pod-1.service", "unit-1.service"},
			map[string]string{
				"pod-private-pod-1.service": "active",
				"unit-1.service":            "active",
			}))
		assert.Equal(t, map[string]*allocation.Header{
			"pod-1": {
				Name:      "pod-1",
				Namespace: "private",
				PodMark:   1,
				AgentMark: 0,
			},
		}, evaluator.List())
	})
	t.Run("destroy pod-1", func(t *testing.T) {
		evaluator.Submit("pod-1", nil)
		time.Sleep(time.Second)
		assert.NoError(t, assertUnits(
			[]string{"pod-private-pod-1.service", "unit-1.service"},
			map[string]string{}))
	})

	assert.NoError(t, evaluator.Close())
	assert.NoError(t, evaluator.Wait())
}
