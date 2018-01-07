// +build ide test_unit

package allocation_test

import (
	"bytes"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSpecMeta_Marshal(t *testing.T) {
	var buf bytes.Buffer
	assert.NoError(t, (&allocation.Spec{
		Revision: "1.3",
	}).Marshal(&buf))
	assert.Equal(t, "### SOIL {\"Revision\":\"1.3\"}\n", buf.String())
}

func TestSpecMeta_Unmarshal(t *testing.T) {
	t.Run(`1.3`, func(t *testing.T) {
		var spec allocation.Spec
		assert.NoError(t, (&spec).Unmarshal("### SOIL {\"Revision\":\"1.3\"}\n"))
		assert.Equal(t, allocation.Spec{
			Revision: "1.3",
		}, spec)
	})
	t.Run(`none`, func(t *testing.T) {
		var spec allocation.Spec
		assert.NoError(t, (&spec).Unmarshal("### POD {}\n"))
		assert.Equal(t, allocation.Spec{
			Revision: "",
		}, spec)
	})
}
