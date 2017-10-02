// +build ide test_unit

package allocation_test

import (
	"testing"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/stretchr/testify/assert"
)

func TestState_FromFS(t *testing.T) {
	paths := allocation.SystemDPaths{
		Local:   "testdata/etc",
		Runtime: "testdata",
	}
	var state allocation.State
	err := (&state).FromFS(paths, "testdata/pod-test-1.service")
	assert.NoError(t, err)
	assert.Len(t, state, 1)
}
