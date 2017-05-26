package scheduler_test

import (
	"context"
	"fmt"
	"github.com/akaspin/concurrency"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/fixture"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"github.com/coreos/go-systemd/dbus"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func assertUnits(names []string, states map[string]string) (err error)  {
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

func TestRestoreAllocation(t *testing.T) {
	sd := fixture.NewSystemd("/run/systemd/system", "pod")
	defer sd.Cleanup()
	assert.NoError(t, sd.DeployPod("test-1", 2))

	alloc, err := scheduler.NewAllocationFromSystemD("/run/systemd/system/pod-test-1.service")
	assert.NoError(t, err)
	assert.Equal(t, &scheduler.Allocation{
		AllocationHeader: &scheduler.AllocationHeader{
			Name:      "test-1",
			PodMark:   123,
			AgentMark: 456,
			Namespace: "private",
		},
		AllocationFile: &scheduler.AllocationFile{
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
		Units: []*scheduler.AllocationUnit{
			{
				AllocationFile: &scheduler.AllocationFile{
					Path: "/run/systemd/system/test-1-0.service",
					Source: `[Unit]
Description=Unit test-1-0.service
[Service]
ExecStart=/usr/bin/sleep inf
[Install]
WantedBy=multi-user.target
`,
				},
				AllocationUnitHeader: &scheduler.AllocationUnitHeader{
					Transition: manifest.Transition{
						Create:  "start",
						Update:  "restart",
						Destroy: "stop",
					},
					Permanent: true,
				},
			},
			{
				AllocationFile: &scheduler.AllocationFile{
					Path: "/run/systemd/system/test-1-1.service",
					Source: `[Unit]
Description=Unit test-1-1.service
[Service]
ExecStart=/usr/bin/sleep inf
[Install]
WantedBy=multi-user.target
`,
				},
				AllocationUnitHeader: &scheduler.AllocationUnitHeader{
					Transition: manifest.Transition{
						Create:  "start",
						Update:  "restart",
						Destroy: "stop",
					},
					Permanent: true,
				},
			},
		},
	}, alloc)
}

func TestNewRuntime(t *testing.T) {
	sd := fixture.NewSystemd("/run/systemd/system", "pod")
	defer sd.Cleanup()
	assert.NoError(t, sd.DeployPod("test-1", 3))
	assert.NoError(t, sd.DeployPod("test-2", 3))

	ctx := context.Background()
	wp := concurrency.NewWorkerPool(ctx, concurrency.Config{
		Capacity: 2,
	})
	ex := scheduler.NewExecutor(ctx, logx.GetLog("test"), wp,)

	sv := supervisor.NewChain(ctx, wp, ex)
	assert.NoError(t, sv.Open())

	assert.NoError(t, sv.Close())
	assert.NoError(t, sv.Wait())
}

func TestRuntime_Submit(t *testing.T) {
	conn, err := dbus.New()
	assert.NoError(t, err)
	defer conn.Close()

	sd := fixture.NewSystemd("/run/systemd/system", "pod")
	defer sd.Cleanup()

	ctx := context.Background()
	wp := concurrency.NewWorkerPool(ctx, concurrency.Config{
		Capacity: 2,
	})
	ex := scheduler.NewExecutor(ctx, logx.GetLog("test"), wp)

	sv := supervisor.NewChain(ctx, wp, ex)
	assert.NoError(t, sv.Open())

	t.Run("create pod-1", func(t *testing.T) {
		alloc := &scheduler.Allocation{
			AllocationHeader: &scheduler.AllocationHeader{
				Name:      "pod-1",
				PodMark:   1,
				AgentMark: 0,
				Namespace: "private",
			},
			AllocationFile: &scheduler.AllocationFile{
				Path: "/run/systemd/system/pod-private-pod-1.service",
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
			Units: []*scheduler.AllocationUnit{
				{
					AllocationFile: &scheduler.AllocationFile{
						Path: "/run/systemd/system/unit-1.service",
						Source: `
[Unit]
Description=Unit 1
[Service]
ExecStart=/usr/bin/sleep inf
[Install]
WantedBy=default.target
`,
					},
					AllocationUnitHeader: &scheduler.AllocationUnitHeader{
						Permanent: false,
						Transition: manifest.Transition{
							Create:  "start",
							Update:  "restart",
							Destroy: "stop",
						},
					},
				},
			},
		}
		ex.Submit("pod-1", alloc)
		time.Sleep(time.Second)

		assert.NoError(t, assertUnits(
			[]string{"pod-private-pod-1.service", "unit-1.service"},
			map[string]string{
				"pod-private-pod-1.service": "active",
				"unit-1.service": "active",
			}))
		assert.Equal(t, map[string]*scheduler.AllocationHeader{
			"pod-1": {
				Name: "pod-1",
				Namespace: "private",
				PodMark: 1,
				AgentMark: 0,
			},
		}, ex.List())
	})
	t.Run("destroy non-existent", func(t *testing.T) {
		ex.Submit("pod-2", nil)
		time.Sleep(time.Second)
		assert.NoError(t, assertUnits(
			[]string{"pod-private-pod-1.service", "unit-1.service"},
			map[string]string{
				"pod-private-pod-1.service": "active",
				"unit-1.service": "active",
			}))
		assert.Equal(t, map[string]*scheduler.AllocationHeader{
			"pod-1": {
				Name: "pod-1",
				Namespace: "private",
				PodMark: 1,
				AgentMark: 0,
			},
		}, ex.List())
	})
	t.Run("destroy pod-1", func(t *testing.T) {
		ex.Submit("pod-1", nil)
		time.Sleep(time.Second)
		assert.NoError(t, assertUnits(
			[]string{"pod-private-pod-1.service", "unit-1.service"},
			map[string]string{}))
		assert.Equal(t, map[string]*scheduler.AllocationHeader{}, ex.List())
	})

	//
	assert.NoError(t, sv.Close())
	assert.NoError(t, sv.Wait())
}
