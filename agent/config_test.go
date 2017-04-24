package agent_test

import (
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_Unmarshal(t *testing.T) {
	config := agent.DefaultConfig()

	assert.Len(t, config.Read("testdata/config1.hcl", "testdata/config2.hcl", "testdata/config3.json", "testdata/non-exists.hcl"), 1)

	assert.Equal(t, &agent.Config{
		Workers: 4,
		Id:   "localhost-1",
		Exec: "ExecStart=/usr/bin/sleep inf",
		Meta: map[string]string{
			"consul":        "true",
			"consul-client": "true",
			"field":         "all,consul",
			"override":      "true",
			"from_json":     "true",
		},
		Local: []*manifest.Pod{
			{
				Name:    "one-1",
				Target:  "default.target",
				Runtime: true,
				Count:   1,
				Constraint: map[string]string{
					"${meta.consul}": "true",
				},
				Units: []*manifest.Unit{
					{
						Name: "one-1-0.service",
						Transition: manifest.Transition{
							Create:  "start",
							Destroy: "stop",
						},
						Source: "[Unit]\nDescription=%p\n\n[Service]\nExecStart=/usr/bin/sleep inf\n\n[Install]\nWantedBy=default.target\n",
					},
				},
			},
			{
				Name:    "one-2",
				Target:  "default.target",
				Runtime: false,
				Count:   0,
				Constraint: map[string]string{
					"${meta.override}": "true",
				},
				Units: []*manifest.Unit{
					{
						Name: "one-2-0.service",
						Transition: manifest.Transition{
							Create:  "start",
							Update:  "restart",
							Destroy: "stop",
						},
						Source: "[Unit]\nDescription=%p\n\n[Service]\nExecStart=/usr/bin/sleep inf\n\n[Install]\nWantedBy=default.target\n",
					},
				},
			},
		},
	}, config)
}
