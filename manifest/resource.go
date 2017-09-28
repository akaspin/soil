package manifest

import (
	"fmt"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
)

// Resources are referenced by ${resource.<Type>.<pod>.name}
type Resource struct {
	Name     string `hcl:"-"`
	Type     string `hcl:"-"`
	Required bool
	Config   map[string]interface{} `hcl:"-"`
}

func defaultResource() (r *Resource) {
	r = &Resource{
		Required: true,
	}
	return
}

func (r *Resource) Id(podName string) (res string) {
	res = fmt.Sprintf("resource.%s.%s.%s", r.Type, podName, r.Name)
	return
}

func (r *Resource) GetConstraint(podName string) (res Constraint) {
	res = Constraint{}
	if r.Required {
		res[fmt.Sprintf("${%s.allocated}", r.Id(podName))] = "true"
	}
	return
}

func (r *Resource) parseAst(raw *ast.ObjectItem) (err error) {
	if len(raw.Keys) != 2 {
		err = fmt.Errorf(`resource should be "type" "name"`)
		return
	}
	r.Type = raw.Keys[0].Token.Value().(string)
	r.Name = raw.Keys[1].Token.Value().(string)
	if err = hcl.DecodeObject(r, raw); err != nil {
		return
	}
	if err = hcl.DecodeObject(&r.Config, raw.Val); err != nil {
		return
	}
	delete(r.Config, "required")
	delete(r.Config, "name")
	delete(r.Config, "type")
	return
}
