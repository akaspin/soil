// +build ide test_unit

package allocation_test

import (
	"bytes"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProviders(t *testing.T) {
	expect := allocation.Providers{
		{
			Nature: "test",
			Kind:   "test",
			Config: map[string]interface{}{
				"a": float64(1),
				"b": `aa "bb"`,
			},
		},
		{
			Nature: "test",
			Kind:   "test2",
			Config: map[string]interface{}{},
		},
	}
	src := "### SOIL provider {\"Nature\":\"test\",\"Kind\":\"test\",\"Config\":{\"a\":1,\"b\":\"aa \\\"bb\\\"\"}}\n### SOIL provider {\"Nature\":\"test\",\"Kind\":\"test2\",\"Config\":{}}\n"
	t.Run(`restore`, func(t *testing.T) {
		var v allocation.Providers
		err := allocation.Recover(&v, &allocation.Provider{}, src, []string{"### SOIL provider"})
		assert.NoError(t, err)
		assert.Equal(t, expect, v)
	})
}

func TestProvider(t *testing.T) {
	expect := &allocation.Provider{
		Nature: "test",
		Kind:   "test",
		Config: map[string]interface{}{
			"a": float64(1),
			"b": `aa "bb"`,
		},
	}
	line := "### SOIL provider {\"Nature\":\"test\",\"Kind\":\"test\",\"Config\":{\"a\":1,\"b\":\"aa \\\"bb\\\"\"}}\n"
	t.Run(`store`, func(t *testing.T) {
		buf := &bytes.Buffer{}
		err := expect.StoreState(buf)
		assert.NoError(t, err)
		assert.Equal(t, line, buf.String())
	})
	t.Run(`restore`, func(t *testing.T) {
		v := &allocation.Provider{}
		err := v.RestoreState(line)
		assert.NoError(t, err)
		assert.Equal(t, expect, v)
	})
}
