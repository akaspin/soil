package manifest

import (
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"strings"
)

type Units []Unit

func (u *Units) Empty() ObjectParser {
	return &Unit{
		Transition: Transition{
			Create:  "start",
			Update:  "restart",
			Destroy: "stop",
		},
	}
}

func (u *Units) Append(v interface{}) (err error) {
	v1 := v.(*Unit)
	*u = append(*u, *v1)
	return
}

type Unit struct {
	Transition `json:",omitempty" hcl:",squash"`
	Name       string
	Source     string
}

func (u Unit) GetID(parent ...string) string {
	return strings.Join(append(parent, u.Name), ".")
}

func (u *Unit) ParseAST(raw *ast.ObjectItem) (err error) {
	u.Name = raw.Keys[0].Token.Value().(string)
	err = hcl.DecodeObject(u, raw)
	u.Source = Heredoc(u.Source)
	return
}

// Unit transition
type Transition struct {
	Create    string
	Update    string
	Destroy   string
	Permanent bool
}
