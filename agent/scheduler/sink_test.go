package scheduler_test

import (
	"context"
	"github.com/akaspin/concurrency"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/metadata"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/fixture"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestScheduler(t *testing.T) {
	pods := []*manifest.Pod{
		{
			Namespace: "private",
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
			Namespace: "private",
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

	// Build supervisor chain

	workerPool := concurrency.NewWorkerPool(ctx, concurrency.Config{
		Capacity: 2,
	})
	executor := scheduler.NewExecutor(ctx, log, workerPool)

	arbiter1 := metadata.NewMapMetadata(ctx, log, "meta", true)
	arbiter2 := metadata.NewMapMetadata(ctx, log, "agent", true)

	// Both map arbiters must be pre initialised
	arbiter1.Configure(map[string]string{
		"consul": "true",
		"test": "true",
	})
	arbiter2.Configure(map[string]string{
		"id": "one",
		"pod_exec": "ExecStart=/usr/bin/sleep inf",
	})

	manager := scheduler.NewArbiter(ctx, log, arbiter1, arbiter2)
	sink := scheduler.NewSink(ctx, logx.GetLog("test"), executor, manager)

	sv := supervisor.NewChain(ctx,
		supervisor.NewChain(ctx,
			supervisor.NewGroup(ctx, arbiter1, arbiter2),
			manager,
		),
		supervisor.NewChain(ctx,
			workerPool,
			executor,
		),
		sink,
	)
	assert.NoError(t, sv.Open())


	t.Run("first sync", func(t *testing.T) {
		sink.Sync("private", pods)
		time.Sleep(time.Second)

		assert.Equal(t, map[string]*scheduler.AllocationHeader{
			"pod-2": {
				Name: "pod-2",
				PodMark: 17758535129175298412,
				AgentMark: 17181089283766246249,
				Namespace: "private",
			},
		}, executor.List())
		time.Sleep(time.Second)
	})
	t.Run("enable pod-3", func(t *testing.T) {
		arbiter1.Configure(map[string]string{
			"consul": "true",
			"test": "true",
			"undefined": "true",
		})

		assert.Equal(t, map[string]*scheduler.AllocationHeader{
			"pod-2": {
				Name: "pod-2",
				PodMark: 17758535129175298412,
				AgentMark: 7974185893712093708,
				Namespace: "private",
			},
			"pod-3": {
				Name: "pod-3",
				PodMark: 5122382858169432775,
				AgentMark: 7974185893712093708,
				Namespace: "private",
			},
		}, executor.List())
		time.Sleep(time.Second)
	})


	assert.NoError(t, sv.Close())
	assert.NoError(t, sv.Wait())

}
