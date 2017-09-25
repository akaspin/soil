package manifest

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"io"
	"os"
)

type Registry []*Pod

func (r *Registry) UnmarshalFiles(namespace string, paths ...string) (err error) {
	var failures []error
	for _, path := range paths {
		failures = append(failures, func(configPath string) (err error) {
			f, err := os.Open(configPath)
			if err != nil {
				return
			}
			defer f.Close()
			err = r.Unmarshal(namespace, f)
			return
		}(path))
	}
	var filtered []error
	for _, failure := range failures {
		if failure != nil {
			filtered = append(filtered, failure)
		}
	}
	if len(filtered) > 0 {
		err = fmt.Errorf("%v", filtered)
	}
	return
}

func (r *Registry) Unmarshal(namespace string, reader io.Reader) (err error) {
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

	pods, err := parseFromAST(namespace, list)
	*r = append(*r, pods...)
	return
}


// Verify registry
func (r *Registry) Verify() (err error) {
	namespaces := map[string]struct{}{}
	for _, pod := range *r {
		namespaces[pod.Namespace] = struct{}{}
	}
	if len(namespaces) > 1 {
		err = fmt.Errorf("multiple namespaces in registry: %v", namespaces)
	}
	return
}
