package scheduler_test

import (
	"context"
	"fmt"
	"github.com/akaspin/concurrency"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/agent/source"
	"github.com/akaspin/soil/fixture"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"github.com/coreos/go-systemd/dbus"
	"github.com/stretchr/testify/assert"
	"reflect"
	"sync"
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

func TestNewExecutor(t *testing.T) {
	sd := fixture.NewSystemd("/run/systemd/system", "pod")
	defer sd.Cleanup()
	assert.NoError(t, sd.DeployPod("test-1", 3))
	assert.NoError(t, sd.DeployPod("test-2", 3))

	ctx := context.Background()
	wp := concurrency.NewWorkerPool(ctx, concurrency.Config{
		Capacity: 2,
	})
	statusReporter := source.NewStatus(ctx, logx.GetLog("test"))
	ex := scheduler.NewExecutor(ctx, logx.GetLog("test"), wp, statusReporter)

	res := map[string]string{}
	var count int
	mu := &sync.Mutex{}
	callback := func(active bool, v map[string]string) {
		mu.Lock()
		defer mu.Unlock()
		if active {
			count++
		}
		res = v
	}

	sv := supervisor.NewChain(ctx, statusReporter, wp, ex)
	assert.NoError(t, sv.Open())
	time.Sleep(time.Second)
	statusReporter.Register(callback)
	time.Sleep(time.Second)
	assert.NoError(t, sv.Close())
	assert.NoError(t, sv.Wait())
	assert.Equal(t, count, 1)
	assert.Equal(t, res, map[string]string{
		"test-2.mark":       "123",
		"test-2.agent_mark": "456",
		"test-2.failures":   "[]",
		"test-1":            "present",
		"test-2":            "present",
		"test-2.namespace":  "private",
		"test-1.namespace":  "private",
		"test-1.mark":       "123",
		"test-1.agent_mark": "456",
		"test-1.failures":   "[]",
		"test-2.units":      "test-2-0.service,test-2-1.service,test-2-2.service",
		"test-1.units":      "test-1-0.service,test-1-1.service,test-1-2.service",
	})
}

func TestExecutor_Submit(t *testing.T) {
	conn, err := dbus.New()
	assert.NoError(t, err)
	defer conn.Close()

	sd := fixture.NewSystemd("/run/systemd/system", "pod")
	defer sd.Cleanup()

	ctx := context.Background()
	wp := concurrency.NewWorkerPool(ctx, concurrency.Config{
		Capacity: 2,
	})
	statusReporter := source.NewStatus(ctx, logx.GetLog("test"))
	ex := scheduler.NewExecutor(ctx, logx.GetLog("test"), wp, statusReporter)
	res := map[string]string{}
	var count int
	mu := &sync.Mutex{}
	callback := func(active bool, v map[string]string) {
		mu.Lock()
		defer mu.Unlock()
		if active {
			count++
		}
		res = v
	}

	sv := supervisor.NewChain(ctx, statusReporter, wp, ex)
	assert.NoError(t, sv.Open())
	statusReporter.Register(callback)

	time.Sleep(time.Second)
	assert.Equal(t, count, 1)
	assert.Equal(t, res, map[string]string{})

	t.Run("create pod-1", func(t *testing.T) {
		alloc := &allocation.Pod{
			Header: &allocation.Header{
				Name:      "pod-1",
				PodMark:   1,
				AgentMark: 0,
				Namespace: "private",
			},
			UnitFile: &allocation.UnitFile{
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
			Units: []*allocation.Unit{
				{
					UnitFile: &allocation.UnitFile{
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
					Transition: &manifest.Transition{
						Permanent: false,
						Create:    "start",
						Update:    "restart",
						Destroy:   "stop",
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
				"unit-1.service":            "active",
			}))
		assert.Equal(t, map[string]*allocation.Header{
			"pod-1": {
				Name:      "pod-1",
				Namespace: "private",
				PodMark:   1,
				AgentMark: 0,
			},
		}, ex.List())
		assert.Equal(t, count, 2)
		assert.Equal(t, res, map[string]string{
			"pod-1.namespace":  "private",
			"pod-1.mark":       "1",
			"pod-1.agent_mark": "0",
			"pod-1.failures":   "[]",
			"pod-1":            "present",
			"pod-1.units":      "unit-1.service",
		})
	})
	t.Run("destroy non-existent", func(t *testing.T) {
		ex.Submit("pod-2", nil)
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
		}, ex.List())
		assert.Equal(t, count, 2)
		assert.Equal(t, res, map[string]string{
			"pod-1.namespace":  "private",
			"pod-1.mark":       "1",
			"pod-1.agent_mark": "0",
			"pod-1.failures":   "[]",
			"pod-1":            "present",
			"pod-1.units":      "unit-1.service",
		})
	})
	t.Run("destroy pod-1", func(t *testing.T) {
		ex.Submit("pod-1", nil)
		time.Sleep(time.Second)
		assert.NoError(t, assertUnits(
			[]string{"pod-private-pod-1.service", "unit-1.service"},
			map[string]string{}))
		assert.Equal(t, map[string]*allocation.Header{}, ex.List())
		assert.Equal(t, count, 3)
		assert.Equal(t, res, map[string]string{})
	})

	//
	assert.NoError(t, sv.Close())
	assert.NoError(t, sv.Wait())
}