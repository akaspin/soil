// +build ide test_unit

package resource2_test

import (
	"github.com/akaspin/soil/agent/resource2"
	"github.com/akaspin/soil/lib"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigs_Unmarshal(t *testing.T) {
	t.Skip()
	var buffers lib.StaticBuffers
	assert.NoError(t, buffers.ReadFiles("testdata/config_test_0.hcl", "testdata/config_test_1.hcl"))
	var configs resource2.Configs
	assert.NoError(t, configs.Unmarshal(buffers.GetReaders()...))
	assert.Equal(t, resource2.Configs{
		resource2.Config{
			Nature: "dummy",
			Kind:   "1",
			Properties: map[string]interface{}{
				"conf_one": 1,
				"conf_two": "two",
			},
		},
		resource2.Config{
			Nature: "range",
			Kind:   "port",
			Properties: map[string]interface{}{
				"max": 9000,
				"min": 8000,
			},
		},
		resource2.Config{
			Nature: "range",
			Kind:   "port_two",
			Properties: map[string]interface{}{
				"min": 9001,
				"max": 10000,
			},
		},
	}, configs)
}