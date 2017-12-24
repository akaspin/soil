package manifest

import (
	"fmt"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/mitchellh/copystructure"
	"strings"
)

const (
	openResourcePrefix    = "resource"
	resourceRequestPrefix = "resource.request"
)

type Resources []Resource

func (r *Resources) Empty() ObjectParser {
	return &Resource{}
}

func (r *Resources) Append(v interface{}) (err error) {
	v1 := v.(*Resource)
	*r = append(*r, *v1)
	return
}

// Resources are referenced by ${resource.<pod>.<name>}
type Resource struct {

	// Resource name unique within pod
	Name string `hcl:"-"`

	// Provider
	Provider string `hcl:"-"`

	// Request config
	Config map[string]interface{} `json:",omitempty" hcl:"-"`
}

func (r Resource) GetID(parent ...string) string {
	return strings.Join(append(parent, r.Name), ".")
}

func (r *Resource) ParseAST(raw *ast.ObjectItem) (err error) {
	if len(raw.Keys) != 2 {
		err = fmt.Errorf(`resource should be "provider" "name"`)
		return
	}
	r.Provider = raw.Keys[0].Token.Value().(string)
	r.Name = raw.Keys[1].Token.Value().(string)
	if err = hcl.DecodeObject(r, raw); err != nil {
		return
	}
	if err = hcl.DecodeObject(&r.Config, raw.Val); err != nil {
		return
	}
	delete(r.Config, "name")
	delete(r.Config, "kind")
	return
}

func (r Resource) Clone() (res Resource) {
	res1, _ := copystructure.Copy(r)
	res = res1.(Resource)
	return
}

// Returns "resource.request.<kind>.allow": "true"
func (r *Resource) GetRequestConstraint() (res Constraint) {
	res = Constraint{
		fmt.Sprintf("${%s.%s.allow}", resourceRequestPrefix, r.Provider): "true",
	}
	return
}

// Returns required constraint for provision with allocated resource
func (r *Resource) GetAllocationConstraint(podName string) (res Constraint) {
	res = Constraint{}
	res[fmt.Sprintf("${%s.%s.%s.allocated}", openResourcePrefix, r.Provider, r.GetID(podName))] = "true"
	return
}

// Returns `resource.<kind>.<pod>.<name>.__values_json`
func (r *Resource) GetValuesKey(podName string) (res string) {
	res = fmt.Sprintf("%s.%s.%s.__values", openResourcePrefix, r.Provider, r.GetID(podName))
	return
}
