// +build ide test_unit

package allocation_test

import (
	"github.com/akaspin/soil/agent/allocation"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestState_FromFS(t *testing.T) {
	paths := allocation.SystemPaths{
		Local:   "testdata/etc",
		Runtime: "testdata",
	}
	var state allocation.State
	err := state.Discover(paths, allocation.GetZeroDiscoveryFunc("testdata/pod-test-1.service"))
	assert.NoError(t, err)
	assert.Len(t, state, 1)
}
