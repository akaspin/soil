package manifest

import (
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/mitchellh/hashstructure"
)

const defaultPodTarget = "default.target"

type Pod struct {
	Name       string
	Runtime    bool
	Target     string
	Count      int
	Constraint map[string]string
	Units      []*Unit
}

func newPodFromItem(raw *ast.ObjectItem) (p *Pod, err error) {
	p = &Pod{
		Target: defaultPodTarget,
		Runtime: true,
	}
	err = hcl.DecodeObject(p, raw)
	p.Name = raw.Keys[0].Token.Value().(string)

	for _, u := range raw.Val.(*ast.ObjectType).List.Filter("unit").Items {
		var unit *Unit
		if unit, err = newUnitFromHCL(u); err != nil {
			return
		}
		p.Units = append(p.Units, unit)
	}
	return
}

func (p *Pod) Mark() (res uint64) {
	res, _ = hashstructure.Hash(p, nil)
	return
}

type Unit struct {
	Transition `hcl:",squash"`
	Name       string
	Permanent  bool
	Source     string
}

func newUnitFromHCL(raw *ast.ObjectItem) (res *Unit, err error) {
	res = &Unit{
		Transition: Transition{
			Destroy: "stop",
		},
	}
	res.Name = raw.Keys[0].Token.Value().(string)
	err = hcl.DecodeObject(res, raw)
	res.Source = Heredoc(res.Source)
	return
}

type Transition struct {
	Create  string
	Update  string
	Destroy string
}
