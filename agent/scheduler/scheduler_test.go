package scheduler_test

import (
	"context"
	"github.com/akaspin/concurrency"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/scheduler/allocation"
	"github.com/akaspin/soil/agent/filter"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/agent/scheduler/executor"
	"github.com/akaspin/soil/fixture"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRuntime_Sync(t *testing.T) {
	pods := []*manifest.Pod{
		{
			Name: "pod-2",
			Runtime: true,
			Target: "default.target",
			Constraint: map[string]string{
				"${meta.consul}": "true",
			},
			Units: []*manifest.Unit{
				{
					Name: "pod-2-unit-1.service",
					Transition: manifest.Transition{
						Create: "start",
						Update: "restart",
						Destroy: "stop",
					},
					Source: `[Unit]
Description=pod-2-unit-1.service
[Service]
ExecStart=/usr/bin/sleep inf
[Install]
WantedBy=default.target
`,
				},
			},
		},
		{
			Name: "pod-3",
			Runtime: true,
			Target: "default.target",
			Constraint: map[string]string{
				"${meta.undefined}": "true",
			},
			Units: []*manifest.Unit{
				{
					Name: "pod-3-unit-1.service",
					Transition: manifest.Transition{
						Create: "start",
						Update: "restart",
						Destroy: "stop",
					},
					Source: `[Unit]
Description=pod-2-unit-1.service
[Service]
ExecStart=/usr/bin/sleep inf
[Install]
WantedBy=default.target
`,
				},
			},
		},
	}

	sd := fixture.NewSystemd("/run/systemd/system", "pod")
	defer sd.Cleanup()
	assert.NoError(t, sd.DeployPod("sync-1", 2))

	time.Sleep(time.Second)

	ctx := context.Background()
	log := logx.GetLog("test")

	// executor
	workerPool := concurrency.NewWorkerPool(ctx, concurrency.Config{
		Capacity: 2,
	})
	executorRt := executor.New(ctx, logx.GetLog("test"), workerPool)
	blockerRt := filter.NewStatic(ctx, log, filter.StaticConfig{
		Id: "one",
		Meta: map[string]string{
			"consul": "true",
			"test": "true",
		},
		PodExec: "ExecStart=/usr/bin/sleep inf",
		Constraint: pods,
	})

	// Scheduler
	schedulerRt := scheduler.NewRuntime(ctx, logx.GetLog("test"), executorRt, blockerRt, "private")

	sv := supervisor.NewChain(ctx, blockerRt, workerPool, executorRt, schedulerRt)
	assert.NoError(t, sv.Open())


	schedulerRt.Sync(pods)
	time.Sleep(time.Second)

	assert.Equal(t, map[string]*allocation.PodHeader{
		"pod-2": {
			Name: "pod-2",
			PodMark: 15470258743007982206,
			AgentMark: 10997408034734681612,
			Namespace: "private",
		},
	}, executorRt.List("private"))

	assert.NoError(t, sv.Close())
	assert.NoError(t, sv.Wait())

}
