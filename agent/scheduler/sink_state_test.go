// +build ide test_unit

package scheduler_test

import (
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/lib"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSchedulerState_SyncNamespace(t *testing.T) {
	state := scheduler.NewSinkState(
		[]string{"private", "public"},
		map[string]string{
			"pod-1": "private",
			"pod-2": "private",
			"pod-3": "public",
			"pod-4": "public",
		},
	)
	t.Run("0 sync private", func(t *testing.T) {
		var buffers lib.StaticBuffers
		var ingest manifest.Pods
		assert.NoError(t, buffers.ReadFiles("testdata/sink_state_test_0.hcl"))
		assert.NoError(t, ingest.Unmarshal(manifest.PrivateNamespace, buffers.GetReaders()...))

		changes := state.SyncNamespace(manifest.PrivateNamespace, ingest)
		assert.Equal(t, map[string]*manifest.Pod{
			"pod-1": ingest[0],
			"pod-2": nil,
		}, changes)
	})
	t.Run("1 sync public", func(t *testing.T) {
		var buffers lib.StaticBuffers
		var ingest manifest.Pods
		assert.NoError(t, buffers.ReadFiles("testdata/sink_state_test_1.hcl"))
		assert.NoError(t, ingest.Unmarshal(manifest.PublicNamespace, buffers.GetReaders()...))

		changes := state.SyncNamespace(manifest.PublicNamespace, ingest)
		assert.Equal(t, map[string]*manifest.Pod{
			"pod-3": ingest[1],
			"pod-4": nil,
		}, changes)
	})
	t.Run("2 remove pod-1 from private", func(t *testing.T) {
		var buffers lib.StaticBuffers
		var ingest manifest.Pods
		assert.NoError(t, buffers.ReadFiles("testdata/sink_state_test_2.hcl"))
		assert.NoError(t, ingest.Unmarshal(manifest.PrivateNamespace, buffers.GetReaders()...))
		changes := state.SyncNamespace(manifest.PrivateNamespace, ingest)
		assert.Len(t, changes, 1)
		assert.Equal(t, changes["pod-1"].Namespace, manifest.PublicNamespace)
	})
	t.Run("3 add pod-1 to private", func(t *testing.T) {
		var buffers lib.StaticBuffers
		var ingest manifest.Pods
		assert.NoError(t, buffers.ReadFiles("testdata/sink_state_test_3.hcl"))
		assert.NoError(t, ingest.Unmarshal(manifest.PrivateNamespace, buffers.GetReaders()...))

		changes := state.SyncNamespace(manifest.PrivateNamespace, ingest)
		assert.Len(t, changes, 1)
		assert.Equal(t, changes["pod-1"].Namespace, manifest.PrivateNamespace)
	})

}
