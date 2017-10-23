// +build ide test_systemd

package provision_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/metrics"
	"github.com/akaspin/soil/agent/provision"
	"github.com/akaspin/soil/fixture"
	"github.com/akaspin/soil/manifest"
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
