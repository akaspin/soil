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
	Agent struct {
		Id string
	}
	Meta   map[string]string `hcl:"meta" json:"meta"`
	System map[string]string `hcl:"system" json:"system"`
}

func DefaultConfig() (c *Config) {
	c = &Config{
		Meta: map[string]string{},
		System: map[string]string{
			"pod_exec": "ExecStart=/usr/bin/sleep inf",
		},
	}
	return
}

func (c *Config) Unmarshal(readers ...io.Reader) (err error) {
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

func (c *Config) unmarshal(r io.Reader) (err error) {
	var buf bytes.Buffer
	if _, err = io.Copy(&buf, r); err != nil {
		return
	}

	root, err := hcl.Parse(buf.String())
	if err != nil {
		return
	}
	buf.Reset()

	list, ok := root.Node.(*ast.ObjectList)
	if !ok {
		err = fmt.Errorf("error parsing: %s", fmt.Errorf("error parsing: root should be an object"))
		return
	}
	var failures []error
	if err = hcl.DecodeObject(c, list); err != nil {
		failures = append(failures, err)
	}
	if len(failures) > 0 {
		err = fmt.Errorf("%v", failures)
	}
	return
}

func (c *Config) Read(path ...string) (err error) {
	var failures []error
	for _, p := range path {
		readErr := func(configPath string) (err error) {
			f, err := os.Open(configPath)
			if err != nil {
				return
			}
			defer f.Close()
			err = c.unmarshal(f)
			return
		}(p)
		if readErr != nil {
			failures = append(failures, readErr)
		}
	}
	if len(failures) > 0 {
		err = fmt.Errorf("%v", failures)
	}
	return
}
