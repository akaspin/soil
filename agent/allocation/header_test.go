// +build ide test_unit

package allocation_test

import (
	"bytes"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHeader_MarshalLine(t *testing.T) {
	var buf bytes.Buffer
	h := allocation.Header{
		Name:      "test",
		AgentMark: 234,
		PodMark:   123,
		Namespace: "private",
	}
	assert.NoError(t, (&h).MarshalSpec(&buf))
	assert.Equal(t, "### POD {\"Name\":\"test\",\"PodMark\":123,\"AgentMark\":234,\"Namespace\":\"private\"}\n", buf.String())
}

func TestHeader_UnmarshalItem(t *testing.T) {
	expect := allocation.Header{
		Name:      "test-1",
		PodMark:   0x7b,
		AgentMark: 0x1c8,
		Namespace: "private",
	}
	t.Run("0", func(t *testing.T) {
		src := `### POD test-1 {"AgentMark":456,"Namespace":"private","PodMark":123}`
		var h allocation.Header
		assert.NoError(t, (&h).UnmarshalSpec(
			src,
			allocation.Spec{
				Revision: "",
			},
			allocation.SystemPaths{}))
		assert.Equal(t, expect, h)
	})
	t.Run("1.0", func(t *testing.T) {
		src := `### POD {"Name":"test-1","AgentMark":456,"Namespace":"private","PodMark":123}`
		var h allocation.Header
		assert.NoError(t, (&h).UnmarshalSpec(
			src,
			allocation.Spec{
				Revision: allocation.SpecRevision,
			},
			allocation.SystemPaths{}))
		assert.Equal(t, expect, h)
	})
}
