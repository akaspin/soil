// +build ide test_systemd

package scheduler_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/fixture"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewScheduler(t *testing.T) {
	sd := fixture.NewSystemd("/run/systemd/system", "pod")
	defer sd.Cleanup()

	ctx := context.Background()
	log := logx.GetLog("test")

	t.Run("0", func(t *testing.T) {

		manager := scheduler.NewManager(ctx, log,
			scheduler.NewManagerSource("agent", false, nil, "private", "public"),
			scheduler.NewManagerSource("meta", false, nil, "private", "public"),
			scheduler.NewManagerSource("system", false, nil, "private", "public"),
		)
		agentSource := bus.NewFlatMap(ctx, log, true, "agent", manager)
		metaSource := bus.NewFlatMap(ctx, log, true, "meta", manager)
		systemSource := bus.NewFlatMap(ctx, log, true, "system", manager)

		executor := scheduler.NewEvaluator(ctx, log)
		sink := scheduler.NewSink(ctx, log, executor, manager)
		sv := supervisor.NewChain(ctx,
			supervisor.NewChain(ctx,
				supervisor.NewGroup(ctx, executor, manager),
				sink,
			),
			supervisor.NewGroup(ctx, agentSource, metaSource, systemSource),
		)
		assert.NoError(t, sv.Open())

		// premature init arbiters
		metaSource.Set(map[string]string{
			"first_private":  "1",
			"second_private": "1",
			"third_public":   "1",
		})
		agentSource.Set(map[string]string{
			"id":       "one",
			"drain":    "false",
		})
		systemSource.Set(map[string]string{
			"pod_exec": "ExecStart=/usr/bin/sleep inf",
		})

		var private, public manifest.Registry
		err := private.UnmarshalFiles("private", "testdata/scheduler_test_0_private.hcl")
		assert.NoError(t, err)
		err = public.UnmarshalFiles("public", "testdata/scheduler_test_0_public.hcl")
		assert.NoError(t, err)

		sink.ConsumeRegistry("private", private)
		sink.ConsumeRegistry("public", public)

		time.Sleep(time.Second)

		res, err := sd.ListPods()
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{
			"first":  "/run/systemd/system/pod-private-first.service",
			"second": "/run/systemd/system/pod-private-second.service",
			"third":  "/run/systemd/system/pod-public-third.service",
		}, res)

		assert.NoError(t, sv.Close())
		assert.NoError(t, sv.Wait())
	})

	// create new arbiter
	manager := scheduler.NewManager(ctx, log,
		scheduler.NewManagerSource("agent", false, nil, "private", "public"),
		scheduler.NewManagerSource("meta", false, nil, "private", "public"),
		scheduler.NewManagerSource("system", false, nil, "private", "public"),
	)
	agentSource := bus.NewFlatMap(ctx, log, true, "agent", manager)
	metaSource := bus.NewFlatMap(ctx, log, true, "meta", manager)
	systemSource := bus.NewFlatMap(ctx, log, true, "system", manager)

	executor := scheduler.NewEvaluator(ctx, log)
	sink := scheduler.NewSink(ctx, log, executor, manager)
	sv := supervisor.NewChain(ctx,
		supervisor.NewChain(ctx,
			supervisor.NewGroup(ctx, executor, manager),
			sink,
		),
		supervisor.NewGroup(ctx, agentSource, metaSource, systemSource),
	)
	assert.NoError(t, sv.Open())

	// premature init arbiters
	assert.NoError(t, sv.Open())
	metaSource.Set(map[string]string{
		"first_private":  "1",
		"second_private": "1",
	})
	agentSource.Set(map[string]string{
		"id":       "one",
	})
	systemSource.Set(map[string]string{
		"pod_exec": "ExecStart=/usr/bin/sleep inf",
	})

	t.Run("1", func(t *testing.T) {
		// assert all pods are still running
		res, err := sd.ListPods()
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{
			"first":  "/run/systemd/system/pod-private-first.service",
			"second": "/run/systemd/system/pod-private-second.service",
			"third":  "/run/systemd/system/pod-public-third.service",
		}, res)
	})
	t.Run("2", func(t *testing.T) {
		var private, public manifest.Registry
		err := private.UnmarshalFiles("private", "testdata/scheduler_test_2_private.hcl")
		assert.NoError(t, err)
		err = public.UnmarshalFiles("public", "testdata/scheduler_test_2_public.hcl")
		assert.NoError(t, err)

		// re sync private
		sink.ConsumeRegistry("private", private)
		time.Sleep(time.Second)

		// Deploy first in public namespace
		sink.ConsumeRegistry("public", public)
		time.Sleep(time.Second)

		// assert first pod is not overrided by public
		// assert third pod is gone
		res, err := sd.ListPods()
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{
			"first":  "/run/systemd/system/pod-private-first.service",
			"second": "/run/systemd/system/pod-private-second.service",
		}, res)
	})
	t.Run("3", func(t *testing.T) {
		// Remove first private
		var private manifest.Registry
		err := private.UnmarshalFiles("private", "testdata/scheduler_test_3_private.hcl")
		assert.NoError(t, err)

		sink.ConsumeRegistry("private", private)
		time.Sleep(time.Second)

		// ensure first is gone
		res, err := sd.ListPods()
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{
			"second": "/run/systemd/system/pod-private-second.service",
		}, res)
	})
	t.Run("4", func(t *testing.T) {
		// modify meta
		metaSource.Set(map[string]string{
			"first_private":  "1",
			"first_public":   "1",
			"second_private": "1",
		})
		time.Sleep(time.Second)

		// ensure first public is deployed
		res, err := sd.ListPods()
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{
			"first":  "/run/systemd/system/pod-public-first.service",
			"second": "/run/systemd/system/pod-private-second.service",
		}, res)
	})
	t.Run("5", func(t *testing.T) {
		// reenter first private
		var private manifest.Registry
		err := private.UnmarshalFiles("private", "testdata/scheduler_test_5_private.hcl")
		assert.NoError(t, err)

		sink.ConsumeRegistry("private", private)
		time.Sleep(time.Second)

		// ensure first is private
		res, err := sd.ListPods()
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{
			"first":  "/run/systemd/system/pod-private-first.service",
			"second": "/run/systemd/system/pod-private-second.service",
		}, res)
	})
	t.Run("6", func(t *testing.T) {
		// remove first private
		var private manifest.Registry
		err := private.UnmarshalFiles("private", "testdata/scheduler_test_6_private.hcl")
		assert.NoError(t, err)

		sink.ConsumeRegistry("private", private)
		time.Sleep(time.Second)

		// ensure first is public
		res, err := sd.ListPods()
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{
			"first":  "/run/systemd/system/pod-public-first.service",
			"second": "/run/systemd/system/pod-private-second.service",
		}, res)
	})
	t.Run("7", func(t *testing.T) {
		// update private and meta
		var private manifest.Registry
		err := private.UnmarshalFiles("private", "testdata/scheduler_test_7_private.hcl")

		assert.NoError(t, err)
		sink.ConsumeRegistry("private", private)

		metaSource.Set(map[string]string{
			"first_private":  "2",
			"first_public":   "1",
			"second_private": "1",
		})
		time.Sleep(time.Second)

		// ensure first is public
		res, err := sd.ListPods()
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{
			"second": "/run/systemd/system/pod-private-second.service",
		}, res)
	})

	assert.NoError(t, sv.Close())
	assert.NoError(t, sv.Wait())
}
