// build ide test_unit

package agent_test

import (
	"github.com/akaspin/soil/agent"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_Unmarshal(t *testing.T) {
	t.Run("no-error", func(t *testing.T) {
		config := agent.DefaultConfig()
		config.Agent.Id = "2"
		assert.NoError(t, config.Read(
			"testdata/config1.hcl",
			"testdata/config2.hcl",
			"testdata/config3.hcl",
			"testdata/config3.json",
		))
		assert.Equal(t, &agent.Config{
			Agent: struct{ Id string }{Id: "1"},
			System: map[string]string{
				"pod_exec": "ExecStart=/usr/bin/sleep inf",
			},
			Meta: map[string]string{
				"consul":        "true",
				"consul-client": "true",
				"field":         "all,consul",
				"override":      "true",
				"from_json":     "true",
				"from-line1":    "true",
				"from-line2":    "true",
			},
		}, config)

	})
	t.Run("non-exists", func(t *testing.T) {
		config := agent.DefaultConfig()
		assert.Error(t, config.Read(
			"testdata/config1.hcl",
			"testdata/config2.hcl",
			"testdata/config3.hcl",
			"testdata/config3.json",
			"testdata/non-exists.hcl",
		))
		assert.Equal(t, &agent.Config{
			Agent: struct{ Id string }{Id: "1"},
			System: map[string]string{
				"pod_exec": "ExecStart=/usr/bin/sleep inf",
			},
			Meta: map[string]string{
				"consul":        "true",
				"consul-client": "true",
				"field":         "all,consul",
				"override":      "true",
				"from_json":     "true",
				"from-line1":    "true",
				"from-line2":    "true",
			},
		}, config)
	})
}
