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
	return nil
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
		return fmt.Errorf(`resource should be "provider" "name"`)
	}
	r.Provider = raw.Keys[0].Token.Value().(string)
	r.Name = raw.Keys[1].Token.Value().(string)
	if err = hcl.DecodeObject(r, raw); err != nil {
		return err
	}
	if err = hcl.DecodeObject(&r.Config, raw.Val); err != nil {
		return err
	}
	delete(r.Config, "name")
	delete(r.Config, "kind")
	return nil
}

func (r Resource) Clone() (res Resource) {
	res1, _ := copystructure.Copy(r)
	return res1.(Resource)
}
