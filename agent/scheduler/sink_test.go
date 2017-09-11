// +build ide test_systemd

package scheduler_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/agent/source"
	"github.com/akaspin/soil/fixture"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSink(t *testing.T) {
	pods := []*manifest.Pod{
		{
			Namespace: "private",
			Name:      "pod-2",
			Runtime:   true,
			Target:    "default.target",
			Constraint: map[string]string{
				"${meta.consul}": "true",
			},
			Units: []*manifest.Unit{
				{
					Name: "pod-2-unit-1.service",
					Transition: manifest.Transition{
						Create:  "start",
						Update:  "restart",
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
			Name:      "pod-3",
			Runtime:   true,
			Target:    "default.target",
			Constraint: map[string]string{
				"${meta.undefined}": "true",
			},
			Units: []*manifest.Unit{
				{
					Name: "pod-3-unit-1.service",
					Transition: manifest.Transition{
						Create:  "start",
						Update:  "restart",
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


	source1 := source.NewPlain(ctx, log, "meta", true)
	source2 := source.NewPlain(ctx, log, "agent", true)
	allocSrc := source.NewAllocation(ctx, log)

	evaluator := scheduler.NewEvaluator(ctx, log, allocSrc)

	// Both map arbiters must be pre initialised
	source1.Configure(map[string]string{
		"consul": "true",
		"test":   "true",
	})
	source2.Configure(map[string]string{
		"id":       "one",
		"pod_exec": "ExecStart=/usr/bin/sleep inf",
	})

	manager := scheduler.NewArbiter(ctx, log, source1, source2, allocSrc)
	sink := scheduler.NewSink(ctx, logx.GetLog("test"), evaluator, manager)

	sv := supervisor.NewChain(ctx,
		supervisor.NewChain(ctx,
			supervisor.NewGroup(ctx, source1, source2),
			manager,
		),
		evaluator,
		sink,
	)
	assert.NoError(t, sv.Open())

	t.Run("first sync", func(t *testing.T) {
		sink.Sync("private", pods)
		time.Sleep(time.Second)

		assert.Equal(t, map[string]*allocation.Header{
			"pod-2": {
				Name:      "pod-2",
				PodMark:   11887013412892795164,
				AgentMark: 17231133757460468042,
				Namespace: "private",
			},
		}, evaluator.List())
		time.Sleep(time.Second)
	})
	t.Run("enable pod-3", func(t *testing.T) {
		source1.Configure(map[string]string{
			"consul":    "true",
			"test":      "true",
			"undefined": "true",
		})

		assert.Equal(t, map[string]*allocation.Header{
			"pod-2": {
				Name:      "pod-2",
				PodMark:   11887013412892795164,
				AgentMark: 14562539397153910086,
				Namespace: "private",
			},
			"pod-3": {
				Name:      "pod-3",
				PodMark:   7050032075987695032,
				AgentMark: 14562539397153910086,
				Namespace: "private",
			},
		}, evaluator.List())
		time.Sleep(time.Second)
	})

	assert.NoError(t, sv.Close())
	assert.NoError(t, sv.Wait())

}
