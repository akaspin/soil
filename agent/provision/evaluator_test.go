// +build ide test_systemd

package provision_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/metrics"
	"github.com/akaspin/soil/agent/provision"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/fixture"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestEvaluator_GetState(t *testing.T) {
	sd := fixture.NewSystemd("/run/systemd/system", "pod")
	defer sd.Cleanup()
	assert.NoError(t, sd.DeployPod("test-1", 3))
	assert.NoError(t, sd.DeployPod("test-2", 3))

	ctx := context.Background()
	reporter := metrics.NewDummy("test")

	systemPaths := allocation.DefaultSystemPaths()
	var state allocation.State
	assert.NoError(t, state.Discover(systemPaths, allocation.DefaultDbusDiscoveryFunc))

	evaluator := provision.NewEvaluator(ctx, logx.GetLog("test"), systemPaths, state, reporter)
	assert.NoError(t, evaluator.Open())
	time.Sleep(time.Second)

	assert.Len(t, reporter.Data, 0)

	state = evaluator.GetState()
	assert.Len(t, state, 2)
	assert.Equal(t, state.Find("test-1"), &allocation.Header{
		Name:      "test-1",
		PodMark:   0x7b,
		AgentMark: 0x1c8,
		Namespace: "private",
	})
	assert.Equal(t, state.Find("test-2"), &allocation.Header{
		Name:      "test-2",
		PodMark:   0x7b,
		AgentMark: 0x1c8,
		Namespace: "private",
	})

	assert.NoError(t, evaluator.Close())
	assert.NoError(t, evaluator.Wait())
}

func TestEvaluator_Allocate(t *testing.T) {
	sd := fixture.NewSystemd("/run/systemd/system", "pod-private")
	sd.Cleanup()
	defer sd.Cleanup()

	ctx := context.Background()
	reporter := metrics.NewDummy("test")

	var state allocation.State
	assert.NoError(t, state.Discover(allocation.DefaultSystemPaths(), allocation.DefaultDbusDiscoveryFunc))

	evaluator := provision.NewEvaluator(ctx, logx.GetLog("test"), allocation.DefaultSystemPaths(), state, reporter)
	assert.NoError(t, evaluator.Open())

	time.Sleep(time.Millisecond * 500)

	t.Run("0 create pod-1", func(t *testing.T) {
		var registry manifest.Registry
		err := registry.UnmarshalFiles("private", "testdata/evaluator_test_Allocate_0.hcl")
		assert.NoError(t, err)

		evaluator.Allocate("pod-1", registry[0], map[string]string{
			"system.pod_exec": "ExecStart=/usr/bin/sleep inf",
		})
		time.Sleep(time.Millisecond * 500)

		sd.AssertUnits(t,
			[]string{"pod-private-pod-1.service", "unit-1.service"},
			map[string]string{
				"pod-private-pod-1.service": "active",
				"unit-1.service":            "active",
			})
		assert.Equal(t, evaluator.GetState().Find("pod-1"),
			&allocation.Header{
				Name:      "pod-1",
				PodMark:   0x2e20553fc77ff55f,
				AgentMark: 0xca6eebba95f6926b,
				Namespace: "private",
			})
		assert.Equal(t, map[string]interface{}{
			"count:provision.evaluations:[]": int64(1),
			"count:provision.failures:[]":    int64(0),
		}, reporter.Data)
	})
	t.Run("1 update pod-1", func(t *testing.T) {
		var registry manifest.Registry
		err := registry.UnmarshalFiles("private", "testdata/evaluator_test_Allocate_1.hcl")
		assert.NoError(t, err)
		evaluator.Allocate("pod-1", registry[0], map[string]string{
			"system.pod_exec": "ExecStart=/usr/bin/sleep inf",
		})
		time.Sleep(time.Millisecond * 500)
		sd.AssertUnits(t,
			[]string{"pod-private-pod-1.service", "unit-1.service"},
			map[string]string{
				"pod-private-pod-1.service": "active",
				"unit-1.service":            "active",
			})
		assert.Equal(t, evaluator.GetState().Find("pod-1"),
			&allocation.Header{
				Name:      "pod-1",
				PodMark:   0x5625baf03ad69cfa,
				AgentMark: 0xca6eebba95f6926b,
				Namespace: "private",
			})
		assert.Equal(t, map[string]interface{}{
			"count:provision.evaluations:[]": int64(2),
			"count:provision.failures:[]":    int64(0),
		}, reporter.Data)
	})
	t.Run("2 destroy non-existent", func(t *testing.T) {
		evaluator.Deallocate("pod-2")
		time.Sleep(time.Millisecond * 500)
		sd.AssertUnits(t,
			[]string{"pod-private-pod-1.service", "unit-1.service"},
			map[string]string{
				"pod-private-pod-1.service": "active",
				"unit-1.service":            "active",
			})
		assert.Equal(t, map[string]interface{}{
			"count:provision.evaluations:[]": int64(2),
			"count:provision.failures:[]":    int64(0),
		}, reporter.Data)
	})
	t.Run("3 destroy pod-1", func(t *testing.T) {
		evaluator.Deallocate("pod-1")
		time.Sleep(time.Millisecond * 500)
		sd.AssertUnits(t, []string{"pod-private-pod-1.service", "unit-1.service"},
			map[string]string{})
		assert.Equal(t, map[string]interface{}{
			"count:provision.evaluations:[]": int64(3),
			"count:provision.failures:[]":    int64(0),
		}, reporter.Data)
	})

	assert.NoError(t, evaluator.Close())
	assert.NoError(t, evaluator.Wait())
}

func TestEvaluator_SinkRestart(t *testing.T) {
	sd := fixture.NewSystemd("/run/systemd/system", "pod")
	defer sd.Cleanup()

	ctx := context.Background()
	log := logx.GetLog("test")
	reporter := metrics.NewDummy("test")

	runOnce := func(t *testing.T) {
		t.Helper()
		manager := scheduler.NewManager(ctx, log, "test",
			scheduler.NewManagerSource("agent", false, nil, "private", "public"),
			scheduler.NewManagerSource("meta", false, nil, "private", "public"),
			scheduler.NewManagerSource("system", false, nil, "private", "public"),
		)
		agentSource := bus.NewStrictMapUpstream("agent", manager)
		metaSource := bus.NewStrictMapUpstream("meta", manager)
		systemSource := bus.NewStrictMapUpstream("system", manager)

		var state allocation.State
		assert.NoError(t, state.Discover(allocation.DefaultSystemPaths(), allocation.DefaultDbusDiscoveryFunc))
		evaluator := provision.NewEvaluator(ctx, log, allocation.DefaultSystemPaths(), state, reporter)
		sink := scheduler.NewSink(ctx, log, state,
			scheduler.NewManagedEvaluator(manager, evaluator))
		sv := supervisor.NewChain(ctx,
			supervisor.NewChain(ctx,
				supervisor.NewGroup(ctx, evaluator, manager),
				sink,
			),
		)
		assert.NoError(t, sv.Open())

		// premature init arbiters
		metaSource.Set(map[string]string{
			"first_private":  "1",
			"second_private": "1",
			"third_public":   "1",
		})
		agentSource.Set(map[string]string{
			"id":    "one",
			"drain": "false",
		})
		systemSource.Set(map[string]string{
			"pod_exec": "ExecStart=/usr/bin/sleep inf",
		})

		var private, public manifest.Registry
		err := private.UnmarshalFiles("private", "testdata/evaluator_test_SinkRestart_0_private.hcl")
		assert.NoError(t, err)
		err = public.UnmarshalFiles("public", "testdata/evaluator_test_SinkRestart_0_public.hcl")
		assert.NoError(t, err)

		sink.ConsumeRegistry("private", private)
		sink.ConsumeRegistry("public", public)

		time.Sleep(time.Millisecond * 1000)

		sd.AssertUnits(t,
			[]string{
				"pod-private-first.service",
				"pod-private-second.service",
				"pod-public-third.service",
				"first-1.service",
				"second-1.service",
				"third-1.service",
			},
			map[string]string{
				"pod-private-first.service":  "active",
				"pod-private-second.service": "active",
				"pod-public-third.service":   "active",
				"first-1.service":            "active",
				"second-1.service":           "active",
				"third-1.service":            "active",
			},
		)

		assert.NoError(t, sv.Close())
		assert.NoError(t, sv.Wait())
	}
	t.Run("0 first run", func(t *testing.T) {
		runOnce(t)
		assert.Equal(t, reporter.Data["count:provision.evaluations:[]"], int64(3))
	})
	t.Run("1 second run", func(t *testing.T) {
		runOnce(t)
		assert.Equal(t, reporter.Data["count:provision.evaluations:[]"], int64(3))
	})
}

func TestEvaluator_SinkFlow(t *testing.T) {
	sd := fixture.NewSystemd("/run/systemd/system", "pod")
	defer sd.Cleanup()

	ctx := context.Background()
	log := logx.GetLog("test")
	reporter := metrics.NewDummy("test")

	manager := scheduler.NewManager(ctx, log, "test",
		scheduler.NewManagerSource("agent", false, nil, "private", "public"),
		scheduler.NewManagerSource("meta", false, nil, "private", "public"),
		scheduler.NewManagerSource("system", false, nil, "private", "public"),
	)
	agentSource := bus.NewStrictMapUpstream("agent", manager)
	metaSource := bus.NewStrictMapUpstream("meta", manager)
	systemSource := bus.NewStrictMapUpstream("system", manager)

	var state allocation.State
	assert.NoError(t, state.Discover(allocation.DefaultSystemPaths(), allocation.DefaultDbusDiscoveryFunc))

	evaluator := provision.NewEvaluator(ctx, log,allocation.DefaultSystemPaths(), state, reporter)
	sink := scheduler.NewSink(ctx, log, state,
		scheduler.NewManagedEvaluator(manager, evaluator))
	sv := supervisor.NewChain(ctx,
		supervisor.NewChain(ctx,
			supervisor.NewGroup(ctx, evaluator, manager),
			sink,
		),
	)
	assert.NoError(t, sv.Open())
	metaSource.Set(map[string]string{
		"first_private":  "1",
		"second_private": "1",
		"third_public":   "1",
	})
	agentSource.Set(map[string]string{
		"id":    "one",
		"drain": "false",
	})
	systemSource.Set(map[string]string{
		"pod_exec": "ExecStart=/usr/bin/sleep inf",
	})

	allUnitNames := []string{
		"pod-private-first.service",
		"pod-public-first.service",
		"pod-private-second.service",
		"pod-public-third.service",
		"first-1.service",
		"second-1.service",
		"third-1.service",
	}

	t.Run("0 deploy private", func(t *testing.T) {
		var registry manifest.Registry
		err := registry.UnmarshalFiles(manifest.PrivateNamespace, "testdata/evaluator_test_SinkFlow_0.hcl")
		assert.NoError(t, err)
		sink.ConsumeRegistry(manifest.PrivateNamespace, registry)
		time.Sleep(time.Millisecond * 500)
		sd.AssertUnits(t, allUnitNames,
			map[string]string{
				"pod-private-first.service":  "active",
				"pod-private-second.service": "active",
				"first-1.service":            "active",
				"second-1.service":           "active",
			})
	})
	t.Run("1 deploy public", func(t *testing.T) {
		var registry manifest.Registry
		err := registry.UnmarshalFiles(manifest.PublicNamespace, "testdata/evaluator_test_SinkFlow_1.hcl")
		assert.NoError(t, err)
		sink.ConsumeRegistry(manifest.PublicNamespace, registry)
		time.Sleep(time.Millisecond * 500)
		sd.AssertUnits(t, allUnitNames,
			map[string]string{
				"pod-private-first.service":  "active",
				"pod-private-second.service": "active",
				"pod-public-third.service":   "active",
				"first-1.service":            "active",
				"second-1.service":           "active",
				"third-1.service":            "active",
			})
	})
	t.Run("2 change constraints of public third", func(t *testing.T) {
		var registry manifest.Registry
		err := registry.UnmarshalFiles(manifest.PublicNamespace, "testdata/evaluator_test_SinkFlow_2.hcl")
		assert.NoError(t, err)
		sink.ConsumeRegistry(manifest.PublicNamespace, registry)
		time.Sleep(time.Millisecond * 500)
		sd.AssertUnits(t, allUnitNames,
			map[string]string{
				"pod-private-first.service":  "active",
				"pod-private-second.service": "active",
				"first-1.service":            "active",
				"second-1.service":           "active",
			})
	})
	t.Run("3 remove private first", func(t *testing.T) {
		var registry manifest.Registry
		err := registry.UnmarshalFiles(manifest.PrivateNamespace, "testdata/evaluator_test_SinkFlow_3.hcl")
		assert.NoError(t, err)
		sink.ConsumeRegistry(manifest.PrivateNamespace, registry)
		time.Sleep(time.Millisecond * 500)
		sd.AssertUnits(t, allUnitNames,
			map[string]string{
				"pod-private-second.service": "active",
				"second-1.service":           "active",
			})
	})
	t.Run("4 add first_public to meta", func(t *testing.T) {
		metaSource.Set(map[string]string{
			"first_private":  "1",
			"first_public":   "1",
			"second_private": "1",
		})
		time.Sleep(time.Millisecond * 500)
		sd.AssertUnits(t, allUnitNames,
			map[string]string{
				"pod-private-second.service": "active",
				"second-1.service":           "active",
				"pod-public-first.service":   "active",
				"first-1.service":            "active",
			})
	})
	t.Run("5 add private first", func(t *testing.T) {
		var registry manifest.Registry
		err := registry.UnmarshalFiles(manifest.PrivateNamespace, "testdata/evaluator_test_SinkFlow_5.hcl")
		assert.NoError(t, err)
		sink.ConsumeRegistry(manifest.PrivateNamespace, registry)
		time.Sleep(time.Millisecond * 500)
		sd.AssertUnits(t, allUnitNames,
			map[string]string{
				"pod-private-second.service": "active",
				"second-1.service":           "active",
				"pod-private-first.service":  "active",
				"first-1.service":            "active",
			})
	})
	t.Run("6 change first_private in meta", func(t *testing.T) {
		metaSource.Set(map[string]string{
			"first_private":  "2",
			"first_public":   "1",
			"second_private": "1",
		})
		time.Sleep(time.Millisecond * 500)
		sd.AssertUnits(t, allUnitNames,
			map[string]string{
				"pod-private-second.service": "active",
				"second-1.service":           "active",
			})
	})
	t.Run("7 remove private first from registry", func(t *testing.T) {
		var registry manifest.Registry
		err := registry.UnmarshalFiles(manifest.PrivateNamespace, "testdata/evaluator_test_SinkFlow_7.hcl")
		assert.NoError(t, err)
		sink.ConsumeRegistry(manifest.PrivateNamespace, registry)
		time.Sleep(time.Millisecond * 500)
		sd.AssertUnits(t, allUnitNames,
			map[string]string{
				"pod-private-second.service": "active",
				"second-1.service":           "active",
				"pod-public-first.service":   "active",
				"first-1.service":            "active",
			})
	})
	t.Run("8 simulate reload with changed registry and meta", func(t *testing.T) {
		metaSource.Set(map[string]string{
			"first_private":  "1",
			"first_public":   "1",
			"second_private": "1",
		})
		var registry manifest.Registry
		err := registry.UnmarshalFiles(manifest.PrivateNamespace, "testdata/evaluator_test_SinkFlow_8.hcl")
		assert.NoError(t, err)
		sink.ConsumeRegistry(manifest.PrivateNamespace, registry)
		time.Sleep(time.Millisecond * 500)
		sd.AssertUnits(t, allUnitNames,
			map[string]string{
				"pod-private-second.service": "active",
				"second-1.service":           "active",
				"pod-private-first.service":  "active",
				"first-1.service":            "active",
			})
	})
	assert.NoError(t, sv.Close())
	assert.NoError(t, sv.Wait())
}
