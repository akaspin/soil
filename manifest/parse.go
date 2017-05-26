package manifest

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"io"
	"os"
)

func ParseFromFiles(namespace string, paths ...string) (res []*Pod, err error) {
	var failures []error
	for _, path := range paths {
		failures = append(failures, func(configPath string) (errs []error) {
			f, err := os.Open(configPath)
			if err != nil {
				errs = append(errs, err)
				return
			}
			defer f.Close()

			var buf bytes.Buffer
			if _, err = io.Copy(&buf, f); err != nil {
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

			pods, errs := ParseFromList(namespace, list)
			res = append(res, pods...)
			return
		}(path)...)
	}
	if len(failures) > 0 {
		err = fmt.Errorf("%v", failures)
	}
	return
}

// Parse manifests from root
func ParseFromReader(namespace string, r io.Reader) (res []*Pod, failures []error, err error) {
	var buf bytes.Buffer
	if _, err = io.Copy(&buf, r); err != nil {
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
		err = fmt.Errorf("error parsing: root should be an object")
		return
	}
	res, failures = ParseFromList(namespace, list)

	return
}

func ParseFromList(namespace string, list *ast.ObjectList) (res []*Pod, failures []error) {
	matches := list.Filter("pod")
	if len(matches.Items) == 0 {
		return
	}

	for _, m := range matches.Items {
		var p *Pod
		var pErr error
		if p, pErr = newPodFromItem(namespace, m); pErr != nil {
			failures = append(failures, pErr)
		}
		res = append(res, p)
	}
	return
}
