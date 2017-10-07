package manifest

import (
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/mitchellh/hashstructure"
	"fmt"
)

const (
	defaultPodTarget = "multi-user.target"
	PrivateNamespace = "private"
	PublicNamespace  = "public"
)

type Pod struct {
	Namespace  string
	Name       string
	Runtime    bool
	Target     string
	Constraint Constraint
	Units      []*Unit
	Blobs      []*Blob
	Resources  []*Resource
}

func DefaultPod(namespace string) (p *Pod) {
	p = &Pod{
		Namespace: namespace,
		Target:    defaultPodTarget,
		Runtime:   true,
	}
	return
}

func (p *Pod) parseAst(raw *ast.ObjectItem) (err error) {
	err = hcl.DecodeObject(p, raw)
	p.Name = raw.Keys[0].Token.Value().(string)

	for _, f := range raw.Val.(*ast.ObjectType).List.Filter("unit").Items {
		unit := defaultUnit()
		if err = unit.parseAst(f); err != nil {
			return
		}
		p.Units = append(p.Units, unit)
	}
	for _, f := range raw.Val.(*ast.ObjectType).List.Filter("blob").Items {
		blob := defaultBlob()
		if err = blob.parseAst(f); err != nil {
			return
		}
		p.Blobs = append(p.Blobs, blob)
	}
	for _, f := range raw.Val.(*ast.ObjectType).List.Filter("resource").Items {
		resource := defaultResource()
		if err = resource.parseAst(f); err != nil {
			return
		}
		p.Resources = append(p.Resources, resource)
	}
	return
}

func (p *Pod) Mark() (res uint64) {
	res, _ = hashstructure.Hash(p, nil)
	return
}

// GetResourceRequestConstraint returns constraint for Resource Evaluator
func (p *Pod) GetResourceRequestConstraint() (res Constraint) {
	res = p.Constraint
	if len(p.Resources) == 0 {
		res = res.Merge(Constraint{
			"${__resource.request.allow}": "false",
		})
		return
	}
	requests := []Constraint{
		{fmt.Sprintf("${%s.request.namespace.%s}", ClosedResourcePrefix, p.Namespace): "true"},
	}
	for _, resource := range p.Resources {
		requests = append(requests, resource.GetRequestConstraint())
	}
	res = res.Merge(requests...)
	return
}

// Returns Pod Constraint with additional Resource Allocation constraints
func (p *Pod) GetResourceAllocationConstraint() (res Constraint) {
	res = p.Constraint
	if len(p.Resources) > 0 {
		resourceConstraint := []Constraint{
			//{"__resource.allocate": "true"},
		}
		for _, resource := range p.Resources {
			resourceConstraint = append(resourceConstraint, resource.GetAllocationConstraint(p.Name))
		}
		res = p.Constraint.Merge(resourceConstraint...)
	}
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
