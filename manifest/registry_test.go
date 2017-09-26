// +build ide test_unit

package manifest_test

import (
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestRegistry_Unmarshal(t *testing.T) {
	var pods manifest.Registry
	r, err := os.Open("testdata/example-multi.hcl")
	assert.NoError(t, err)
	defer r.Close()

	err = (&pods).Unmarshal("private", r)
	assert.NoError(t, err)
	assert.Len(t, pods, 2)
}

func TestRegistry_UnmarshalFiles(t *testing.T) {
	var pods manifest.Registry
	err := pods.UnmarshalFiles("private", "testdata/files_1.hcl", "testdata/files_2.hcl")
	assert.NoError(t, err)
	assert.Len(t, pods, 3)
}
