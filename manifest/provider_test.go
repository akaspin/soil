// +build ide test_unit

package manifest_test

import (
	"github.com/akaspin/soil/lib"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestProvider_ParseAST(t *testing.T) {
	t.Run(`no errors`, func(t *testing.T) {
		var buffers lib.StaticBuffers
		require.NoError(t, buffers.ReadFiles("testdata/TestProviders_ParseAST_0.hcl"))
		roots, err := lib.ParseHCL(buffers.GetReaders()...)
		var providers manifest.Providers
		err = manifest.ParseList(roots, "provider", &providers)
		assert.NoError(t, err)
		assert.Equal(t,
			manifest.Providers{
				{Kind: "a", Name: "1", Config: map[string]interface{}{}},
				{Kind: "a", Name: "2", Config: map[string]interface{}{}},
				{Kind: "c", Name: "3", Config: map[string]interface{}{"any": 1}},
			},
			providers)
	})
}
