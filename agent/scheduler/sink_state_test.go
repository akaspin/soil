// +build ide test_unit

package scheduler_test

import (
	"github.com/akaspin/soil/agent/scheduler"
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
	t.Run("private", func(t *testing.T) {
		ingest, err := manifest.ParseFromFiles("private", "testdata/sink_state_test_1.hcl")
		assert.NoError(t, err)
		changes := state.SyncNamespace("private", ingest)
		assert.Equal(t, map[string]*manifest.Pod{
			"pod-2": nil,
			"pod-1": ingest[0],
		}, changes)
	})
	t.Run("public", func(t *testing.T) {
		//ingestPrivate, err := manifest.ParseFromFiles("private", "testdata/sink_state_test_1.hcl")
		//assert.NoError(t, err)

		ingestPublic, err := manifest.ParseFromFiles("public", "testdata/sink_state_test_2.hcl")
		assert.NoError(t, err)
		changes := state.SyncNamespace("public", ingestPublic)
		assert.Equal(t, map[string]*manifest.Pod{
			//"pod-1": ingestPrivate[0],
			"pod-4": nil,
			"pod-3": ingestPublic[1],
		}, changes)
	})

}
