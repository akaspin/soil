package manifest

import (
	"fmt"
	"github.com/hashicorp/hcl/hcl/ast"
)

func parseFromAST(namespace string, list *ast.ObjectList) (res []*Pod, err error) {
	matches := list.Filter("pod")
	if len(matches.Items) == 0 {
		return
	}

	var failures []error
	for _, m := range matches.Items {
		p := DefaultPod(namespace)
		var pErr error
		if pErr = p.parseAst(m); pErr != nil {
			failures = append(failures, pErr)
		}
		res = append(res, p)
	}
	if len(failures) > 0 {
		err = fmt.Errorf("%v", failures)
	}
	return
}
