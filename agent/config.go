package agent

import (
	"bytes"
	"fmt"
	"github.com/akaspin/soil/manifest"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"io"
	"os"
)

type Config struct {
	Id string
	Meta map[string]string `hcl:"meta" json:"meta"`
	Exec string
	Workers int
	Local []*manifest.Pod
}

func DefaultConfig() (c *Config)  {
	c = &Config{
		Workers: 4,
		Id: "localhost",
		Meta: map[string]string{},
		Exec: "ExecStart=/usr/bin/sleep inf",
	}
	return
}

func (c *Config) Unmarshal(r io.Reader) (failures []error) {
	var err error
	var buf bytes.Buffer
	if _, err = io.Copy(&buf, r); err != nil {
		failures = append(failures, err)
		return
	}

	root, err := hcl.Parse(buf.String())
	if err != nil {
		failures = append(failures, fmt.Errorf("error parsing: %s", err))
		return
	}
	buf.Reset()

	list, ok := root.Node.(*ast.ObjectList)
	if !ok {
		failures = append(failures, fmt.Errorf("error parsing: %s", fmt.Errorf("error parsing: root should be an object")))
		return
	}
	for _, chunk := range list.Filter("agent").Items {
		if err = hcl.DecodeObject(c, chunk); err != nil {
			failures = append(failures, err)
		}
	}
	pods, podFailures := manifest.ParseFromList(list)
	c.Local = append(c.Local, pods...)
	failures = append(failures, podFailures...)
	return
}

func (c *Config) Read(path ...string) (failures []error) {
	for _, p := range path {
		failures = append(failures, func(configPath string) (errs []error) {
			f, err := os.Open(configPath)
			if err != nil {
				errs = append(errs, err)
				return
			}
			defer f.Close()
			errs = c.Unmarshal(f)
			return
		}(p)...)
	}
	return
}