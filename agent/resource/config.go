package resource

import "github.com/mitchellh/hashstructure"

// Config represents one resource configuration
type Config struct {
	Type       string
	Name       string
	Properties map[string]interface{}
}

func (c *Config) IsEqual(config *Config) (res bool) {
	leftHash, _ := hashstructure.Hash(c, nil)
	rightHash, _ := hashstructure.Hash(config, nil)
	res = leftHash == rightHash
	return
}