package resource

import "github.com/mitchellh/hashstructure"

// Config represents one resource in Agent configuration
type Config struct {
	Nature     string                 // Worker nature
	Kind       string                 // Declared type
	Properties map[string]interface{} // Properties
}

func (c *Config) IsEqual(config *Config) (res bool) {
	leftHash, _ := hashstructure.Hash(c, nil)
	rightHash, _ := hashstructure.Hash(config, nil)
	res = leftHash == rightHash
	return
}

// Static external configuration propagated to all workers and executors
type EvaluatorConfig struct {
}