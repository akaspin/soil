package manifest

import (
	"fmt"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"strings"
)

type Providers []Provider

func (p *Providers) Empty() ObjectParser {
	return &Provider{}
}

func (p *Providers) Append(v interface{}) (err error) {
	v1 := v.(*Provider)
	*p = append(*p, *v1)
	return
}

// Resource provider
type Provider struct {
	Kind   string                 // Resource kind: range, pool ...
	Name   string                 // Logical name unique within pod
	Config map[string]interface{} `json:",omitempty"`
}

func (p Provider) GetID(parent ...string) string {
	return strings.Join(append(parent, p.Name), ".")
}

func (p Provider) ID(parent string) string {
	return parent + `.` + p.Kind + `.` + p.Name
}

func (p *Provider) ParseAST(raw *ast.ObjectItem) (err error) {
	if len(raw.Keys) != 2 {
		err = fmt.Errorf(`provuder should be "nature" "name"`)
		return
	}
	p.Kind = raw.Keys[0].Token.Value().(string)
	p.Name = raw.Keys[1].Token.Value().(string)
	if err = hcl.DecodeObject(p, raw); err != nil {
		return
	}
	if err = hcl.DecodeObject(&p.Config, raw.Val); err != nil {
		return
	}
	delete(p.Config, "nature")
	delete(p.Config, "kind")
	return
}
