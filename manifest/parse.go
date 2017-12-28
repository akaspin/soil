package manifest

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/hcl/ast"
	"sort"
)

type ListParser interface {
	Empty() ObjectParser
	Append(v interface{}) (err error)
}

type ObjectParser interface {
	WithID
	ParseAST(raw *ast.ObjectItem) (err error)
}

func ParseList(lists []*ast.ObjectList, key string, parser ListParser) (err error) {
	var res1 []ObjectParser
	err = &multierror.Error{}

	for _, list := range lists {
		for _, obj := range list.Filter(key).Items {
			objParser := parser.Empty()
			if parseErr := objParser.ParseAST(obj); parseErr != nil {
				multierror.Append(err, parseErr)
				continue
			}
			res1 = append(res1, objParser)
		}
	}
	sort.Slice(res1, func(i, j int) bool {
		return res1[i].GetID() < res1[j].GetID()
	})
	var lastId = ""
	for _, obj := range res1 {
		if id := obj.GetID(); lastId != id {
			lastId = id
			if appendErr := parser.Append(obj); appendErr != nil {
				multierror.Append(err, appendErr)
			}
		} else {
			multierror.Append(err, fmt.Errorf(`%s with %s already defined`, key, id))
		}
	}
	err = err.(*multierror.Error).ErrorOrNil()
	return
}
