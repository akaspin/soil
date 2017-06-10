package manifest

import (
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/mitchellh/hashstructure"
)

const (
	defaultPodTarget = "multi-user.target"
)

type Pod struct {
	Namespace  string
	Name       string
	Runtime    bool
	Target     string
	Constraint Constraint
	Units      []*Unit
	Blobs      []*Blob
}

func newPodFromItem(namespace string, raw *ast.ObjectItem) (p *Pod, err error) {
	p = &Pod{
		Namespace: namespace,
		Target:    defaultPodTarget,
		Runtime:   true,
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
	for _, f := range raw.Val.(*ast.ObjectType).List.Filter("blob").Items {
		var blob *Blob
		if blob, err = newBlobFromHCL(f); err != nil {
			return
		}
		p.Blobs = append(p.Blobs, blob)
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
	Source     string
}

func newUnitFromHCL(raw *ast.ObjectItem) (res *Unit, err error) {
	res = &Unit{
		Transition: Transition{
			Create:  "start",
			Update:  "restart",
			Destroy: "stop",
		},
	}
	res.Name = raw.Keys[0].Token.Value().(string)
	err = hcl.DecodeObject(res, raw)
	res.Source = Heredoc(res.Source)
	return
}

// Unit transition
type Transition struct {
	Create    string
	Update    string
	Destroy   string
	Permanent bool
}

// Pod file
type Blob struct {
	Name        string
	Permissions int
	Leave       bool
	Source      string
}

func newBlobFromHCL(raw *ast.ObjectItem) (res *Blob, err error) {
	res = &Blob{
		Permissions: 0644,
	}
	res.Name = raw.Keys[0].Token.Value().(string)
	err = hcl.DecodeObject(res, raw)
	res.Source = Heredoc(res.Source)
	return
}

// Compare
func IsEqual(left, right *Pod) (ok bool) {
	if left == nil {
		if right != nil {
			return
		}
		ok = true
		return
	}
	if left.Mark() == right.Mark() {
		ok = true
	}
	return
}
