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
		failures = append(failures, func(configPath string) (err error) {
			f, err := os.Open(configPath)
			if err != nil {
				return
			}
			defer f.Close()

			var buf bytes.Buffer
			if _, err = io.Copy(&buf, f); err != nil {
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
			res = append(res, pods...)
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

func parseFromAST(namespace string, list *ast.ObjectList) (res []*Pod, err error) {
	matches := list.Filter("pod")
	if len(matches.Items) == 0 {
		return
	}

	var failures []error
	for _, m := range matches.Items {
		p := DefaultPod(namespace)
		var pErr error
		if pErr = p.UnmarshalAST(m); pErr != nil {
			failures = append(failures, pErr)
		}
		res = append(res, p)
	}
	if len(failures) > 0 {
		err = fmt.Errorf("%v", failures)
	}
	return
}
