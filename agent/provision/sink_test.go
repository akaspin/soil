// +build ide test_systemd

package provision_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/provision"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/fixture"
	"github.com/akaspin/soil/lib"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestEvaluator_SinkFlow(t *testing.T) {
	sd := fixture.NewSystemd("/run/systemd/system", "pod")
	defer sd.Cleanup()

	ctx := context.Background()
	log := logx.GetLog("test")

	arbiter := scheduler.NewArbiter(ctx, log, "test", scheduler.ArbiterConfig{})
	var state allocation.Recovery
	assert.NoError(t, state.FromFilesystem(allocation.DefaultSystemPaths(), allocation.DefaultDbusDiscoveryFunc))
	evaluator := provision.NewEvaluator(ctx, log, provision.EvaluatorConfig{
		SystemPaths:    allocation.DefaultSystemPaths(),
		Recovery:       state,
		StatusConsumer: &bus.BlackholePipe{},
	})
	sink := scheduler.NewSink(ctx, log, state,
		scheduler.NewBoundedEvaluator(arbiter, evaluator))
	sv := supervisor.NewChain(ctx,
		supervisor.NewChain(ctx,
			supervisor.NewGroup(ctx, evaluator, arbiter),
			sink,
		),
	)
	assert.NoError(t, sv.Open())

	arbiter.ConsumeMessage(bus.NewMessage("", map[string]string{
		"system.pod_exec":     "ExecStart=/usr/bin/sleep inf",
		"meta.first_private":  "1",
		"meta.second_private": "1",
		"meta.third_public":   "1",
	}))

	allUnitNames := []string{
		"pod-private-first.service",
		"pod-public-first.service",
		"pod-private-second.service",
		"pod-public-third.service",
		"first-1.service",
		"second-1.service",
		"third-1.service",
	}

	waitConfig := fixture.WaitConfig{
		Retry:   time.Millisecond * 50,
		Retries: 1000,
	}

	t.Run("0 deploy private", func(t *testing.T) {
		var buffers lib.StaticBuffers
		var registry manifest.Registry
		assert.NoError(t, buffers.ReadFiles("testdata/evaluator_test_SinkFlow_0.hcl"))
		assert.NoError(t, registry.Unmarshal(manifest.PrivateNamespace, buffers.GetReaders()...))

		sink.ConsumeRegistry(registry)
		fixture.WaitNoError(t, waitConfig, sd.UnitStatesFn(allUnitNames,
			map[string]string{
				"pod-private-first.service":  "active",
				"pod-private-second.service": "active",
				"first-1.service":            "active",
				"second-1.service":           "active",
			}))
		sd.AssertUnitHashes(t, allUnitNames,
			map[string]uint64{
				"/run/systemd/system/pod-private-first.service":  0x42552996e6deba52,
				"/run/systemd/system/first-1.service":            0x6ac69815b89bddee,
				"/run/systemd/system/pod-private-second.service": 0x967a3661747e5983,
				"/run/systemd/system/second-1.service":           0x6ac69815b89bddee,
			})
	})
	t.Run("1 deploy public", func(t *testing.T) {
		var buffers lib.StaticBuffers
		var registry manifest.Registry
		assert.NoError(t, buffers.ReadFiles("testdata/evaluator_test_SinkFlow_1.hcl"))
		assert.NoError(t, registry.Unmarshal(manifest.PublicNamespace, buffers.GetReaders()...))

		sink.ConsumeRegistry(registry)
		fixture.WaitNoError(t, waitConfig, sd.UnitStatesFn(allUnitNames,
			map[string]string{
				"pod-private-first.service":  "active",
				"pod-private-second.service": "active",
				"pod-public-third.service":   "active",
				"first-1.service":            "active",
				"second-1.service":           "active",
				"third-1.service":            "active",
			}))

		sd.AssertUnitHashes(t, allUnitNames,
			map[string]uint64{
				"/run/systemd/system/second-1.service":           0x6ac69815b89bddee,
				"/run/systemd/system/pod-private-first.service":  0x42552996e6deba52,
				"/run/systemd/system/pod-private-second.service": 0x967a3661747e5983,
				"/run/systemd/system/pod-public-third.service":   0x7ea2ffac5a0a1b96,
				"/run/systemd/system/third-1.service":            0xdcdd742d1352ae8e,
				"/run/systemd/system/first-1.service":            0x6ac69815b89bddee,
			})
	})
	t.Run("2 change constraints of public third", func(t *testing.T) {
		var buffers lib.StaticBuffers
		var registry manifest.Registry
		assert.NoError(t, buffers.ReadFiles("testdata/evaluator_test_SinkFlow_2.hcl"))
		assert.NoError(t, registry.Unmarshal(manifest.PublicNamespace, buffers.GetReaders()...))

		sink.ConsumeRegistry(registry)
		fixture.WaitNoError(t, waitConfig, sd.UnitStatesFn(allUnitNames,
			map[string]string{
				"pod-private-first.service":  "active",
				"pod-private-second.service": "active",
				"first-1.service":            "active",
				"second-1.service":           "active",
			}))
		sd.AssertUnitHashes(t, allUnitNames,
			map[string]uint64{
				"/run/systemd/system/second-1.service":           0x6ac69815b89bddee,
				"/run/systemd/system/pod-private-first.service":  0x42552996e6deba52,
				"/run/systemd/system/pod-private-second.service": 0x967a3661747e5983,
				"/run/systemd/system/first-1.service":            0x6ac69815b89bddee,
			})
	})
	t.Run("3 remove private first", func(t *testing.T) {
		var buffers lib.StaticBuffers
		var registry manifest.Registry
		assert.NoError(t, buffers.ReadFiles("testdata/evaluator_test_SinkFlow_3.hcl"))
		assert.NoError(t, registry.Unmarshal(manifest.PrivateNamespace, buffers.GetReaders()...))
		sink.ConsumeRegistry(registry)

		fixture.WaitNoError(t, waitConfig, sd.UnitStatesFn(allUnitNames,
			map[string]string{
				"pod-private-second.service": "active",
				"second-1.service":           "active",
			}))
		sd.AssertUnitHashes(t, allUnitNames,
			map[string]uint64{
				"/run/systemd/system/pod-private-second.service": 0x967a3661747e5983,
				"/run/systemd/system/second-1.service":           0x6ac69815b89bddee,
			})
	})
	t.Run("4 add first_public to meta", func(t *testing.T) {
		arbiter.ConsumeMessage(bus.NewMessage("", map[string]string{
			"system.pod_exec":     "ExecStart=/usr/bin/sleep inf",
			"meta.first_private":  "1",
			"meta.second_private": "1",
			"meta.first_public":   "1",
		}))
		fixture.WaitNoError(t, waitConfig, sd.UnitStatesFn(allUnitNames,
			map[string]string{
				"pod-private-second.service": "active",
				"second-1.service":           "active",
				"pod-public-first.service":   "active",
				"first-1.service":            "active",
			}))
		sd.AssertUnitHashes(t, allUnitNames,
			map[string]uint64{
				"/run/systemd/system/pod-private-second.service": 0x918a2f1a57136223,
				"/run/systemd/system/first-1.service":            0x6ac69815b89bddee,
				"/run/systemd/system/pod-public-first.service":   0x73ec639c3434642d,
				"/run/systemd/system/second-1.service":           0x6ac69815b89bddee,
			})
	})
	t.Run("5 add private first to registry", func(t *testing.T) {
		var buffers lib.StaticBuffers
		var registry manifest.Registry
		assert.NoError(t, buffers.ReadFiles("testdata/evaluator_test_SinkFlow_5.hcl"))
		assert.NoError(t, registry.Unmarshal(manifest.PrivateNamespace, buffers.GetReaders()...))
		sink.ConsumeRegistry(registry)

		fixture.WaitNoError(t, waitConfig, sd.UnitStatesFn(allUnitNames,
			map[string]string{
				"pod-private-second.service": "active",
				"second-1.service":           "active",
				"pod-private-first.service":  "active",
				"first-1.service":            "active",
			}))
		sd.AssertUnitHashes(t, allUnitNames,
			map[string]uint64{
				"/run/systemd/system/second-1.service":           0x6ac69815b89bddee,
				"/run/systemd/system/pod-private-first.service":  0x4a119305f1a13473,
				"/run/systemd/system/pod-private-second.service": 0x918a2f1a57136223,
				"/run/systemd/system/first-1.service":            0x6ac69815b89bddee,
			})
	})
	t.Run("6 change first_private in meta", func(t *testing.T) {
		arbiter.ConsumeMessage(bus.NewMessage("", map[string]string{
			"system.pod_exec":     "ExecStart=/usr/bin/sleep inf",
			"meta.first_private":  "2",
			"meta.first_public":   "1",
			"meta.second_private": "1",
		}))

		fixture.WaitNoError(t, waitConfig, sd.UnitStatesFn(allUnitNames,
			map[string]string{
				"pod-private-second.service": "active",
				"second-1.service":           "active",
			}))
		sd.AssertUnitHashes(t, allUnitNames,
			map[string]uint64{
				"/run/systemd/system/pod-private-second.service": 0xb57ff0c462e8e8c4,
				"/run/systemd/system/second-1.service":           0x6ac69815b89bddee,
			})
	})
	t.Run("7 remove private first from registry", func(t *testing.T) {
		var buffers lib.StaticBuffers
		var registry manifest.Registry
		assert.NoError(t, buffers.ReadFiles("testdata/evaluator_test_SinkFlow_7.hcl"))
		assert.NoError(t, registry.Unmarshal(manifest.PrivateNamespace, buffers.GetReaders()...))

		sink.ConsumeRegistry(registry)
		fixture.WaitNoError(t, waitConfig, sd.UnitStatesFn(allUnitNames,
			map[string]string{
				"pod-private-second.service": "active",
				"second-1.service":           "active",
				"pod-public-first.service":   "active",
				"first-1.service":            "active",
			}))
		sd.AssertUnitHashes(t, allUnitNames,
			map[string]uint64{
				"/run/systemd/system/pod-private-second.service": 0xb57ff0c462e8e8c4,
				"/run/systemd/system/pod-public-first.service":   0x43188d38cbda3df3,
				"/run/systemd/system/second-1.service":           0x6ac69815b89bddee,
				"/run/systemd/system/first-1.service":            0x6ac69815b89bddee,
			})
	})
	t.Run("8 simulate reload with changed registry and meta", func(t *testing.T) {
		arbiter.ConsumeMessage(bus.NewMessage("", map[string]string{
			"system.pod_exec":     "ExecStart=/usr/bin/sleep inf",
			"meta.first_private":  "1",
			"meta.first_public":   "1",
			"meta.second_private": "1",
		}))
		var buffers lib.StaticBuffers
		var registry manifest.Registry
		assert.NoError(t, buffers.ReadFiles("testdata/evaluator_test_SinkFlow_8.hcl"))
		assert.NoError(t, registry.Unmarshal(manifest.PrivateNamespace, buffers.GetReaders()...))

		sink.ConsumeRegistry(registry)

		fixture.WaitNoError(t, waitConfig, sd.UnitStatesFn(allUnitNames,
			map[string]string{
				"pod-private-second.service": "active",
				"second-1.service":           "active",
				"pod-private-first.service":  "active",
				"first-1.service":            "active",
			}))
		sd.AssertUnitHashes(t, allUnitNames,
			map[string]uint64{
				"/run/systemd/system/pod-private-first.service":  0x4a119305f1a13473,
				"/run/systemd/system/pod-private-second.service": 0x918a2f1a57136223,
				"/run/systemd/system/second-1.service":           0x6ac69815b89bddee,
				"/run/systemd/system/first-1.service":            0x6ac69815b89bddee,
			})
	})

	assert.NoError(t, sv.Close())
	assert.NoError(t, sv.Wait())
}
