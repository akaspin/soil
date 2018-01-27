package agent

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"io"
	"os"
)

// Agent - specific config
type Config struct {
	Meta   map[string]string `hcl:"meta" json:"meta"`
	System map[string]string `hcl:"system" json:"system"`
}

func DefaultConfig() (c *Config) {
	return &Config{
		Meta: map[string]string{},
		System: map[string]string{
			"pod_exec": "ExecStart=/usr/bin/sleep inf",
		},
	}
}

func (c *Config) Unmarshal(readers ...io.Reader) (err error) {
	var failures []error
	for _, reader := range readers {
		if failure := c.unmarshal(reader); failure != nil {
			failures = append(failures, failure)
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("%v", failures)
	}
	return nil
}

func (c *Config) unmarshal(r io.Reader) (err error) {
	var buf bytes.Buffer
	if _, err = io.Copy(&buf, r); err != nil {
		return err
	}

	root, err := hcl.Parse(buf.String())
	if err != nil {
		return err
	}
	buf.Reset()

	list, ok := root.Node.(*ast.ObjectList)
	if !ok {
		err = fmt.Errorf("error parsing: %s", fmt.Errorf("error parsing: root should be an object"))
		return err
	}
	var failures []error
	if err = hcl.DecodeObject(c, list); err != nil {
		failures = append(failures, err)
	}
	if len(failures) > 0 {
		return fmt.Errorf("%v", failures)
	}
	return nil
}

func (c *Config) Read(path ...string) (err error) {
	var failures []error
	for _, p := range path {
		readErr := func(configPath string) (err error) {
			f, err := os.Open(configPath)
			if err != nil {
				return err
			}
			defer f.Close()
			return c.unmarshal(f)
		}(p)
		if readErr != nil {
			failures = append(failures, readErr)
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("%v", failures)
	}
	return nil
}
