package manifest

import (
	"fmt"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/mitchellh/hashstructure"
	"sort"
	"strings"
)

const defaultPodTarget = "default.target"



type Pod struct {
	Namespace string
	Name       string
	Runtime    bool
	Target     string
	Constraint Constraint
	Units      []*Unit
}

func newPodFromItem(namespace string, raw *ast.ObjectItem) (p *Pod, err error) {
	p = &Pod{
		Namespace: namespace,
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

type Constraint map[string]string

// Extract constraint fields by namespaces
func (c Constraint) ExtractFields() (res map[string][]string) {
	res = map[string][]string{}
	collected := map[string]struct{}{}
	for k, v := range c {
		for _, f := range ExtractEnv(k+v) {
			collected[f] = struct{}{}
		}
	}
	for k := range collected {
		split := strings.SplitN(k, ".", 2)
		if len(split) == 2 {
			res[split[0]] = append(res[split[0]], split[1])
		}
	}
	for _, v := range res {
		sort.Strings(v)
	}
	return
}

func (c Constraint) Check(env map[string]string) (err error) {
	for left, right := range c {
		leftV := Interpolate(left, env)
		rightV := Interpolate(right, env)
		if leftV != rightV {
			err = fmt.Errorf("constraint failed %s != %s (%s:%s)", leftV, rightV, left, right)
			return
		}
	}
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
