package config

import "github.com/akaspin/soil/agent/resource"

// Agent configuration
type AgentConfig struct {
	Meta   map[string]string
	System struct {
		PodExec string `json:"pod_exec" hcl:"pod_exec"`
	}
	Resource []resource.Config
}
