package manifest

import (
	"fmt"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
)

const (
	OpenResourcePrefix   = "resource"
	ClosedResourcePrefix = "__resource"
	ResourceRequestPrefix = "__resource.request"
	ResourceValuesPrefix = "__resource.values"
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

func (r *Resource) GetId(podName string) (res string) {
	res = fmt.Sprintf("%s.%s", podName, r.Name)
	return
}

func (r *Resource) GetRequestConstraint() (res Constraint) {
	res = Constraint{
		fmt.Sprintf("${%s.type.%s}", ResourceRequestPrefix, r.Type): "true",
	}
	return
}

// GetAllocationConstraint returns required constraint for provision with allocated resource
func (r *Resource) GetAllocationConstraint(podName string) (res Constraint) {
	res = Constraint{}
	if r.Required {
		res[fmt.Sprintf("${%s.%s.%s.allocated}", OpenResourcePrefix, r.Type, r.GetId(podName))] = "true"
	}
	return
}

func (r *Resource) GetValuesKey(podName string) (res string) {
	res = fmt.Sprintf("%s.%s.%s", ResourceValuesPrefix, r.Type, r.GetId(podName))
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
