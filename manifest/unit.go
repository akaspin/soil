package manifest

import (
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
)

type Unit struct {
	Transition `hcl:",squash"`
	Name       string
	Source     string
}

func defaultUnit() (u Unit) {
	u = Unit{
		Transition: Transition{
			Create:  "start",
			Update:  "restart",
			Destroy: "stop",
		},
	}
	return
}

func (u *Unit) parseAst(raw *ast.ObjectItem) (err error) {
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
