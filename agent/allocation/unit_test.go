// +build ide test_unit

package allocation_test

import (
	"bytes"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnit_MarshalLine(t *testing.T) {
	u := &allocation.Unit{
		UnitFile: allocation.UnitFile{
			Path: "aaa",
		},
		Transition: manifest.Transition{
			Create: "start",
		},
	}
	var buf bytes.Buffer
	assert.NoError(t, u.MarshalLine(&buf))
	assert.Equal(t, "### BLOB_V2 {\"Path\":\"aaa\",\"Create\":\"start\"}\n", buf.String())
}
