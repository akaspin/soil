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

func TestEvaluator_Allocate(t *testing.T) {
	sd := fixture.NewSystemd("/run/systemd/system", "pod-private")
	sd.Cleanup()
	defer sd.Cleanup()

	ctx := context.Background()

	var state allocation.Recovery
	assert.NoError(t, state.FromFilesystem(allocation.DefaultSystemPaths(), allocation.DefaultDbusDiscoveryFunc))

	evaluator := provision.NewEvaluator(ctx, logx.GetLog("test"), allocation.DefaultSystemPaths(), state, &metrics.BlackHole{})
	assert.NoError(t, evaluator.Open())

	time.Sleep(time.Millisecond * 500)

	t.Run("0 create pod-1", func(t *testing.T) {
		var registry manifest.Registry
		err := registry.UnmarshalFiles("private", "testdata/evaluator_test_Allocate_0.hcl")
		assert.NoError(t, err)

		evaluator.Allocate(registry[0], map[string]string{
			"system.pod_exec": "ExecStart=/usr/bin/sleep inf",
		})
		time.Sleep(time.Millisecond * 500)

		sd.AssertUnitStates(t,
			[]string{"pod-private-pod-1.service", "unit-1.service"},
			map[string]string{
				"pod-private-pod-1.service": "active",
				"unit-1.service":            "active",
			})
		sd.AssertUnitHashes(t,
			[]string{"pod-private-pod-1.service", "unit-1.service"},
			map[string]uint64{
				"/run/systemd/system/pod-private-pod-1.service": 0xc43253a8821be2b,
				"/run/systemd/system/unit-1.service":            0xbca69ea672e79d81,
			},
		)
	})
	t.Run("1 update pod-1", func(t *testing.T) {
		var registry manifest.Registry
		err := registry.UnmarshalFiles("private", "testdata/evaluator_test_Allocate_1.hcl")
		assert.NoError(t, err)
		evaluator.Allocate(registry[0], map[string]string{
			"system.pod_exec": "ExecStart=/usr/bin/sleep inf",
		})
		time.Sleep(time.Millisecond * 500)
		sd.AssertUnitStates(t,
			[]string{"pod-private-pod-1.service", "unit-1.service"},
			map[string]string{
				"pod-private-pod-1.service": "active",
				"unit-1.service":            "active",
			})
		sd.AssertUnitHashes(t,
			[]string{"pod-private-pod-1.service", "unit-1.service"},
			map[string]uint64{
				"/run/systemd/system/pod-private-pod-1.service": 0x28525a605380724b,
				"/run/systemd/system/unit-1.service":            0x448529ac4d4389a0,
			},
		)
	})
	t.Run("2 destroy non-existent", func(t *testing.T) {
		evaluator.Deallocate("pod-2")
		time.Sleep(time.Millisecond * 500)
		sd.AssertUnitStates(t,
			[]string{"pod-private-pod-1.service", "unit-1.service"},
			map[string]string{
				"pod-private-pod-1.service": "active",
				"unit-1.service":            "active",
			})
		sd.AssertUnitHashes(t,
			[]string{"pod-private-pod-1.service", "unit-1.service"},
			map[string]uint64{
				"/run/systemd/system/pod-private-pod-1.service": 0x28525a605380724b,
				"/run/systemd/system/unit-1.service":            0x448529ac4d4389a0,
			},
		)
	})
	t.Run("3 destroy pod-1", func(t *testing.T) {
		evaluator.Deallocate("pod-1")
		time.Sleep(time.Millisecond * 500)
		sd.AssertUnitStates(t, []string{"pod-private-pod-1.service", "unit-1.service"},
			map[string]string{})
		sd.AssertUnitHashes(t,
			[]string{"pod-private-pod-1.service", "unit-1.service"},
			map[string]uint64{},
		)
	})

	assert.NoError(t, evaluator.Close())
	assert.NoError(t, evaluator.Wait())
}

func TestEvaluator_SinkRestart(t *testing.T) {
	sd := fixture.NewSystemd("/run/systemd/system", "pod")
	defer sd.Cleanup()

	ctx := context.Background()
	log := logx.GetLog("test")

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

		var state allocation.Recovery
		assert.NoError(t, state.FromFilesystem(allocation.DefaultSystemPaths(), allocation.DefaultDbusDiscoveryFunc))
		evaluator := provision.NewEvaluator(ctx, log, allocation.DefaultSystemPaths(), state, &metrics.BlackHole{})
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

		var private, public manifest.Registry
		err := private.UnmarshalFiles("private", "testdata/evaluator_test_SinkRestart_0_private.hcl")
		assert.NoError(t, err)
		err = public.UnmarshalFiles("public", "testdata/evaluator_test_SinkRestart_0_public.hcl")
		assert.NoError(t, err)

		sink.ConsumeRegistry("private", private)
		sink.ConsumeRegistry("public", public)

		time.Sleep(time.Millisecond * 1000)

		sd.AssertUnitStates(t,
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
		sd.AssertUnitHashes(t,
			[]string{
				"pod-private-first.service",
				"pod-private-second.service",
				"pod-public-third.service",
				"first-1.service",
				"second-1.service",
				"third-1.service",
			},
			map[string]uint64{
				"/run/systemd/system/pod-private-second.service": 0xd80dbdb828d2560a,
				"/run/systemd/system/third-1.service":            0xdcdd742d1352ae8e,
				"/run/systemd/system/pod-private-first.service":  0x2f1e0cdd2dddc667,
				"/run/systemd/system/second-1.service":           0x6ac69815b89bddee,
				"/run/systemd/system/first-1.service":            0x6ac69815b89bddee,
				"/run/systemd/system/pod-public-third.service":   0x9dd961b3cb2c5880,
			},
		)

		assert.NoError(t, sv.Close())
		assert.NoError(t, sv.Wait())
	}
	t.Run("0 first run", func(t *testing.T) {
		runOnce(t)
	})
	t.Run("1 second run", func(t *testing.T) {
		runOnce(t)
	})
}

func TestEvaluator_SinkFlow(t *testing.T) {
	sd := fixture.NewSystemd("/run/systemd/system", "pod")
	defer sd.Cleanup()

	ctx := context.Background()
	log := logx.GetLog("test")

	manager := scheduler.NewManager(ctx, log, "test",
		scheduler.NewManagerSource("agent", false, nil, "private", "public"),
		scheduler.NewManagerSource("meta", false, nil, "private", "public"),
		scheduler.NewManagerSource("system", false, nil, "private", "public"),
	)
	agentSource := bus.NewStrictMapUpstream("agent", manager)
	metaSource := bus.NewStrictMapUpstream("meta", manager)
	systemSource := bus.NewStrictMapUpstream("system", manager)

	var state allocation.Recovery
	assert.NoError(t, state.FromFilesystem(allocation.DefaultSystemPaths(), allocation.DefaultDbusDiscoveryFunc))

	evaluator := provision.NewEvaluator(ctx, log, allocation.DefaultSystemPaths(), state, &metrics.BlackHole{})
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
		sd.AssertUnitStates(t, allUnitNames,
			map[string]string{
				"pod-private-first.service":  "active",
				"pod-private-second.service": "active",
				"first-1.service":            "active",
				"second-1.service":           "active",
			})
		sd.AssertUnitHashes(t, allUnitNames,
			map[string]uint64{
				"/run/systemd/system/pod-private-first.service":  0x2f1e0cdd2dddc667,
				"/run/systemd/system/pod-private-second.service": 0xd80dbdb828d2560a,
				"/run/systemd/system/first-1.service":            0x6ac69815b89bddee,
				"/run/systemd/system/second-1.service":           0x6ac69815b89bddee,
			})
	})
	t.Run("1 deploy public", func(t *testing.T) {
		var registry manifest.Registry
		err := registry.UnmarshalFiles(manifest.PublicNamespace, "testdata/evaluator_test_SinkFlow_1.hcl")
		assert.NoError(t, err)
		sink.ConsumeRegistry(manifest.PublicNamespace, registry)
		time.Sleep(time.Millisecond * 500)
		sd.AssertUnitStates(t, allUnitNames,
			map[string]string{
				"pod-private-first.service":  "active",
				"pod-private-second.service": "active",
				"pod-public-third.service":   "active",
				"first-1.service":            "active",
				"second-1.service":           "active",
				"third-1.service":            "active",
			})
		sd.AssertUnitHashes(t, allUnitNames,
			map[string]uint64{
				"/run/systemd/system/pod-private-first.service":  0x2f1e0cdd2dddc667,
				"/run/systemd/system/pod-private-second.service": 0xd80dbdb828d2560a,
				"/run/systemd/system/first-1.service":            0x6ac69815b89bddee,
				"/run/systemd/system/second-1.service":           0x6ac69815b89bddee,
				// new
				"/run/systemd/system/pod-public-third.service": 0x9dd961b3cb2c5880,
				"/run/systemd/system/third-1.service":          0xdcdd742d1352ae8e,
			})
	})
	t.Run("2 change constraints of public third", func(t *testing.T) {
		var registry manifest.Registry
		err := registry.UnmarshalFiles(manifest.PublicNamespace, "testdata/evaluator_test_SinkFlow_2.hcl")
		assert.NoError(t, err)
		sink.ConsumeRegistry(manifest.PublicNamespace, registry)
		time.Sleep(time.Millisecond * 500)
		sd.AssertUnitStates(t, allUnitNames,
			map[string]string{
				"pod-private-first.service":  "active",
				"pod-private-second.service": "active",
				"first-1.service":            "active",
				"second-1.service":           "active",
			})
		sd.AssertUnitHashes(t, allUnitNames,
			map[string]uint64{
				"/run/systemd/system/pod-private-first.service":  0x2f1e0cdd2dddc667,
				"/run/systemd/system/pod-private-second.service": 0xd80dbdb828d2560a,
				"/run/systemd/system/first-1.service":            0x6ac69815b89bddee,
				"/run/systemd/system/second-1.service":           0x6ac69815b89bddee,
			})
	})
	t.Run("3 remove private first", func(t *testing.T) {
		var registry manifest.Registry
		err := registry.UnmarshalFiles(manifest.PrivateNamespace, "testdata/evaluator_test_SinkFlow_3.hcl")
		assert.NoError(t, err)
		sink.ConsumeRegistry(manifest.PrivateNamespace, registry)
		time.Sleep(time.Millisecond * 500)
		sd.AssertUnitStates(t, allUnitNames,
			map[string]string{
				"pod-private-second.service": "active",
				"second-1.service":           "active",
			})
		sd.AssertUnitHashes(t, allUnitNames,
			map[string]uint64{
				"/run/systemd/system/pod-private-second.service": 0xd80dbdb828d2560a,
				"/run/systemd/system/second-1.service":           0x6ac69815b89bddee,
			})
	})
	t.Run("4 add first_public to meta", func(t *testing.T) {
		metaSource.Set(map[string]string{
			"first_private":  "1",
			"first_public":   "1",
			"second_private": "1",
		})
		time.Sleep(time.Millisecond * 500)
		sd.AssertUnitStates(t, allUnitNames,
			map[string]string{
				"pod-private-second.service": "active",
				"second-1.service":           "active",
				"pod-public-first.service":   "active",
				"first-1.service":            "active",
			})
		sd.AssertUnitHashes(t, allUnitNames,
			map[string]uint64{
				// pods are changed
				"/run/systemd/system/pod-public-first.service":   0x9f24f379a7845013,
				"/run/systemd/system/pod-private-second.service": 0x99da87008a6d05f2,
				// units are not changed
				"/run/systemd/system/first-1.service":  0x6ac69815b89bddee,
				"/run/systemd/system/second-1.service": 0x6ac69815b89bddee,
			})
	})
	t.Run("5 add private first", func(t *testing.T) {
		var registry manifest.Registry
		err := registry.UnmarshalFiles(manifest.PrivateNamespace, "testdata/evaluator_test_SinkFlow_5.hcl")
		assert.NoError(t, err)
		sink.ConsumeRegistry(manifest.PrivateNamespace, registry)
		time.Sleep(time.Millisecond * 500)
		sd.AssertUnitStates(t, allUnitNames,
			map[string]string{
				"pod-private-second.service": "active",
				"second-1.service":           "active",
				"pod-private-first.service":  "active",
				"first-1.service":            "active",
			})
		sd.AssertUnitHashes(t, allUnitNames,
			map[string]uint64{
				// first pod now is private
				"/run/systemd/system/pod-private-first.service": 0xa067f24f9caeb5d,
				// second pod is not changed
				"/run/systemd/system/pod-private-second.service": 0x99da87008a6d05f2,
				// units are not changed
				"/run/systemd/system/first-1.service":  0x6ac69815b89bddee,
				"/run/systemd/system/second-1.service": 0x6ac69815b89bddee,
			})
	})
	t.Run("6 change first_private in meta", func(t *testing.T) {
		metaSource.Set(map[string]string{
			"first_private":  "2",
			"first_public":   "1",
			"second_private": "1",
		})
		time.Sleep(time.Millisecond * 500)
		sd.AssertUnitStates(t, allUnitNames,
			map[string]string{
				"pod-private-second.service": "active",
				"second-1.service":           "active",
			})
		sd.AssertUnitHashes(t, allUnitNames,
			map[string]uint64{
				// second pod is changed
				"/run/systemd/system/pod-private-second.service": 0x2b61532a72875e85,
				// units are not changed
				"/run/systemd/system/second-1.service": 0x6ac69815b89bddee,
			})
	})
	t.Run("7 remove private first from registry", func(t *testing.T) {
		var registry manifest.Registry
		err := registry.UnmarshalFiles(manifest.PrivateNamespace, "testdata/evaluator_test_SinkFlow_7.hcl")
		assert.NoError(t, err)
		sink.ConsumeRegistry(manifest.PrivateNamespace, registry)
		time.Sleep(time.Millisecond * 500)
		sd.AssertUnitStates(t, allUnitNames,
			map[string]string{
				"pod-private-second.service": "active",
				"second-1.service":           "active",
				"pod-public-first.service":   "active",
				"first-1.service":            "active",
			})
		sd.AssertUnitHashes(t, allUnitNames,
			map[string]uint64{
				"/run/systemd/system/pod-public-first.service":   0x5c593f7c22807a92,
				"/run/systemd/system/pod-private-second.service": 0x2b61532a72875e85,
				"/run/systemd/system/first-1.service":            0x6ac69815b89bddee,
				"/run/systemd/system/second-1.service":           0x6ac69815b89bddee,
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
		sd.AssertUnitStates(t, allUnitNames,
			map[string]string{
				"pod-private-second.service": "active",
				"second-1.service":           "active",
				"pod-private-first.service":  "active",
				"first-1.service":            "active",
			})
		sd.AssertUnitHashes(t, allUnitNames,
			map[string]uint64{
				"/run/systemd/system/pod-private-first.service":  0xa067f24f9caeb5d,
				"/run/systemd/system/pod-private-second.service": 0x99da87008a6d05f2,
				"/run/systemd/system/first-1.service":            0x6ac69815b89bddee,
				"/run/systemd/system/second-1.service":           0x6ac69815b89bddee,
			})
	})
	assert.NoError(t, sv.Close())
	assert.NoError(t, sv.Wait())
}
