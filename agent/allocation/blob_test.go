// +build ide test_unit

package allocation_test

import (
	"bytes"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBlob_UnmarshalItem(t *testing.T) {
	expect := allocation.Blob{
		Name:        "testdata/blob.txt",
		Permissions: 0,
		Leave:       true,
		Source:      "a\nb\n123\n",
	}
	t.Run(`v1`, func(t *testing.T) {
		line := `### BLOB testdata/blob.txt {"Leave":true}`
		var b allocation.Blob
		assert.NoError(t, (&b).UnmarshalItem(line))
		assert.Equal(t, expect, b)
	})
	t.Run(`v2`, func(t *testing.T) {
		line := `### BLOB_V2 {"Name":"testdata/blob.txt","Leave":true}`
		var b allocation.Blob
		assert.NoError(t, (&b).UnmarshalItem(line))
	})
}

func TestBlob_MarshalLine(t *testing.T) {
	b := allocation.Blob{
		Name:        "testdata/blob.txt",
		Permissions: 0,
		Leave:       true,
		Source:      "a\nb\n123\n",
	}
	var buf bytes.Buffer
	assert.NoError(t, b.MarshalLine(&buf))
	assert.Equal(t, "### BLOB_V2 {\"Name\":\"testdata/blob.txt\",\"Leave\":true}\n", buf.String())
}
