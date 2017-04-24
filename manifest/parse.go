package manifest

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"io"
)

// Parse manifests from root
func ParseManifests(r io.Reader) (res []*Pod, failures []error, err error) {
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
	res, failures = ParseFromList(list)

	return
}

func ParseFromList(list *ast.ObjectList) (res []*Pod, failures []error) {
	matches := list.Filter("pod")
	if len(matches.Items) == 0 {
		return
	}

	for _, m := range matches.Items {
		var p *Pod
		var pErr error
		if p, pErr = newPodFromItem(m); pErr != nil {
			failures = append(failures, pErr)
		}
		res = append(res, p)
	}
	return
}
