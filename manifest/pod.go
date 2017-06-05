package manifest

import (
	"fmt"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/mitchellh/hashstructure"
	"sort"
	"strings"
	"strconv"
)

const (
	defaultPodTarget = "default.target"

	opEqual = "="
	opLess = "<"
	opGreater = ">"
	opIn = "~"
)



type Pod struct {
	Namespace string
	Name       string
	Runtime    bool
	Target     string
	Constraint Constraint
	Units      []*Unit
	Files []*File
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
	for _, f := range raw.Val.(*ast.ObjectType).List.Filter("file").Items {
		var blob *File
		if blob, err = newFileFromHCL(f); err != nil {
			return
		}
		p.Files = append(p.Files, blob)
	}
	return
}

func (p *Pod) Mark() (res uint64) {
	res, _ = hashstructure.Hash(p, nil)
	return
}

// Constraint can contain interpolations in form ${ns.key}.
// Right field can also begins with compare operation: "<", ">" or "~" (in).
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
		if !checkPair(leftV, rightV) {
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

// Unit transition
type Transition struct {
	Create  string
	Update  string
	Destroy string
}

// Pod file
type File struct {
	Name string
	Permissions int
	Leave bool
	Source string
}

func newFileFromHCL(raw *ast.ObjectItem) (res *File, err error) {
	res = &File{
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

func checkPair(left, right string) (res bool) {
	// check for operation
	op := opEqual
	split := strings.SplitN(right, " ", 2)
	if len(split) == 2 {
		// have op
		switch split[0] {
		case opLess, opGreater:
			op = split[0]
			right = split[1]
			leftN, leftErr := strconv.ParseFloat(left, 64)
			rightN, rightErr := strconv.ParseFloat(split[1], 64)
			if leftErr != nil || rightErr != nil {
				switch op {
				case opLess:
					res = left < right
				case opGreater:
					res = left > right
				}
			} else {
				switch op {
				case opLess:
					res = leftN < rightN
				case opGreater:
					res = leftN > rightN
				}
			}
			return
		case opIn:
			// inside
			rightSplit := strings.Split(split[1], ",")
			LEFT:
			for _, leftChunk := range strings.Split(left, ",") {
				for _, rightChunk := range rightSplit {
					if strings.TrimSpace(leftChunk) == strings.TrimSpace(rightChunk) {
						continue LEFT
					}
				}
				// nothing found
				return
			}
			// found all
			res = true
		}
		return
	}
	// ordinal string comparison
	res = left == right
	return
}
