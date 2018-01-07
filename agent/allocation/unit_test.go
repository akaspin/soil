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
	assert.Equal(t, "### UNIT_V2 {\"Path\":\"aaa\",\"Create\":\"start\"}\n", buf.String())
}

func TestUnit_UnmarshalItem(t *testing.T) {
	expect := allocation.Unit{
		UnitFile: allocation.UnitFile{
			Path:   "testdata/test-1-0.service",
			Source: "[Unit]\nDescription=Unit test-1-0.service\n[Service]\nExecStart=/usr/bin/sleep inf\n[Install]\nWantedBy=multi-user.target\n"},
		Transition: manifest.Transition{
			Create: "start",
		}}

	t.Run(`v1`, func(t *testing.T) {
		line := `### UNIT testdata/test-1-0.service {"Create":"start"}`
		var u allocation.Unit
		assert.NoError(t, (&u).UnmarshalItem(line, allocation.SystemPaths{}))
		assert.Equal(t, expect, u)
	})
	t.Run(`v2`, func(t *testing.T) {
		line := `### UNIT_V2 {"Path":"testdata/test-1-0.service","Create":"start"}`
		var u allocation.Unit
		assert.NoError(t, (&u).UnmarshalItem(line, allocation.SystemPaths{}))
		assert.Equal(t, expect, u)
	})
}
