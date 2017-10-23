package resource

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/mitchellh/hashstructure"
	"io"
)

// Static external configuration propagated to all workers and executors
type EvaluatorConfig struct {
}

// Config represents one resource in Agent configuration
type Config struct {
	Nature     string                 // Worker nature
	Kind       string                 // Declared type
	Properties map[string]interface{} `hcl:",squash"` // Properties
}

func (c *Config) IsEqual(config Config) (res bool) {
	leftHash, _ := hashstructure.Hash(*c, nil)
	rightHash, _ := hashstructure.Hash(config, nil)
	res = leftHash == rightHash
	return
}

func (c *Config) parseAst(m *ast.ObjectItem) (err error) {
	if len(m.Keys) != 2 {
		err = fmt.Errorf(`resource config should be named as "nature" "name"`)
		return
	}
	if err = hcl.DecodeObject(&(c.Properties), m.Val); err != nil {
		return
	}
	c.Nature = m.Keys[0].Token.Value().(string)
	c.Kind = m.Keys[1].Token.Value().(string)
	return
}

type Configs []Config

func (c *Configs) Unmarshal(readers ...io.Reader) (err error) {
	var failures []error
	for _, reader := range readers {
		if failure := c.unmarshal(reader); failure != nil {
			failures = append(failures, failure)
		}
	}
	if len(failures) > 0 {
		err = fmt.Errorf("%v", failures)
	}
	return
}

func (c *Configs) unmarshal(reader io.Reader) (err error) {
	var buf bytes.Buffer
	if _, err = io.Copy(&buf, reader); err != nil {
		return
	}
	root, err := hcl.Parse(buf.String())
	if err != nil {
		err = fmt.Errorf("error parsing: %s", err)
		return
	}
	buf.Reset()

	list, ok := root.Node.(*ast.ObjectList)
	if !ok {
		err = fmt.Errorf("error parsing: %s", fmt.Errorf("error parsing: root should be an object"))
		return
	}
	matches := list.Filter("resource")

	var failures []error
	for _, m := range matches.Items {
		var config Config
		if failure := config.parseAst(m); failure != nil {
			failures = append(failures, failure)
			continue
		}
		*c = append(*c, config)
	}
	if len(failures) > 0 {
		err = fmt.Errorf("%v", failures)
	}
	return
}
