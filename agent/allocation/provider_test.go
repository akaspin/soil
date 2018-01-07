// +build ide test_unit

package allocation_test

import (
	"bytes"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProviderSlice_Append(t *testing.T) {
	expect := allocation.ProviderSlice{
		{
			Kind: "test",
			Name: "test",
			Config: map[string]interface{}{
				"a": float64(1),
				"b": `aa "bb"`,
			},
		},
		{
			Kind:   "test",
			Name:   "test2",
			Config: map[string]interface{}{},
		},
	}
	src := "### PROVIDER {\"Kind\":\"test\",\"Name\":\"test\",\"Config\":{\"a\":1,\"b\":\"aa \\\"bb\\\"\"}}\n### PROVIDER {\"Kind\":\"test\",\"Name\":\"test2\",\"Config\":{}}\n"
	t.Run(`restore`, func(t *testing.T) {
		var v allocation.ProviderSlice
		var spec allocation.Spec
		err := spec.UnmarshalAssetSlice(allocation.SystemPaths{}, &v, src)
		assert.NoError(t, err)
		assert.Equal(t, expect, v)
	})
}

func TestProvider(t *testing.T) {
	expect := &allocation.Provider{
		Kind: "test",
		Name: "test",
		Config: map[string]interface{}{
			"a": float64(1),
			"b": `aa "bb"`,
		},
	}
	line := "### PROVIDER {\"Kind\":\"test\",\"Name\":\"test\",\"Config\":{\"a\":1,\"b\":\"aa \\\"bb\\\"\"}}\n"
	t.Run(`store`, func(t *testing.T) {
		buf := &bytes.Buffer{}
		err := expect.MarshalSpec(buf)
		assert.NoError(t, err)
		assert.Equal(t, line, buf.String())
	})
	t.Run(`restore`, func(t *testing.T) {
		v := &allocation.Provider{}
		err := v.UnmarshalSpec(line, allocation.Spec{}, allocation.SystemPaths{})
		assert.NoError(t, err)
		assert.Equal(t, expect, v)
	})
}

func TestProviders_FromManifest(t *testing.T) {
	man := manifest.Pod{
		Providers: manifest.Providers{
			{
				Kind: "test",
				Name: "test",
				Config: map[string]interface{}{
					"a": float64(1),
					"b": `aa "bb"`,
				},
			},
			{
				Kind:   "test",
				Name:   "test2",
				Config: map[string]interface{}{},
			},
		},
	}
	var providers allocation.ProviderSlice
	assert.NoError(t, providers.FromManifest(man, nil))
	assert.Equal(t, allocation.ProviderSlice{
		{
			Kind: "test",
			Name: "test",
			Config: map[string]interface{}{
				"a": float64(1),
				"b": `aa "bb"`,
			},
		},
		{
			Kind:   "test",
			Name:   "test2",
			Config: map[string]interface{}{},
		},
	},
		providers)
}
