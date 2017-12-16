package manifest

import (
	"encoding/json"
	"fmt"
	"github.com/akaspin/soil/lib"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"hash/crc64"
	"io"
	"strings"
)

const (
	defaultPodTarget = "multi-user.target"
	PrivateNamespace = "private"
	PublicNamespace  = "public"
)

type Pods []*Pod

func (r *Pods) Empty() ObjectParser {
	return &Pod{
		Namespace: PrivateNamespace,
		Target:    defaultPodTarget,
		Runtime:   true,
	}
}

func (r *Pods) Append(v interface{}) (err error) {
	*r = append(*r, v.(*Pod))
	return
}

func (r *Pods) SetNamespace(namespace string) {
	for _, pod := range *r {
		pod.Namespace = namespace
	}
}

func (r *Pods) Unmarshal(namespace string, reader ...io.Reader) (err error) {
	err = &multierror.Error{}
	roots, parseErr := lib.ParseHCL(reader...)
	err = multierror.Append(err, parseErr)
	err = multierror.Append(err, ParseList(roots, "pod", r))
	r.SetNamespace(namespace)
	err = err.(*multierror.Error).ErrorOrNil()
	return
}

// Pod manifest
type Pod struct {
	Namespace  string
	Name       string
	Runtime    bool
	Target     string
	Constraint Constraint `json:",omitempty"`
	Units      Units      `json:",omitempty" hcl:"-"`
	Blobs      Blobs      `json:",omitempty" hcl:"-"`
	Resources  Resources  `json:",omitempty" hcl:"-"`
	Providers  Providers  `json:",omitempty" hcl:"-"`
}

func (p Pod) GetID(parent ...string) string {
	return strings.Join(append(parent, p.Namespace, p.Name), ".")
}

func (p *Pod) ParseAST(raw *ast.ObjectItem) (err error) {
	err = &multierror.Error{}
	list := raw.Val.(*ast.ObjectType).List

	if err = multierror.Append(err, hcl.DecodeObject(p, raw)); err.(*multierror.Error).ErrorOrNil() != nil {
		return
	}
	p.Name = raw.Keys[0].Token.Value().(string)

	err = multierror.Append(err, ParseList([]*ast.ObjectList{list}, "unit", &p.Units))
	err = multierror.Append(err, ParseList([]*ast.ObjectList{list}, "blob", &p.Blobs))
	err = multierror.Append(err, ParseList([]*ast.ObjectList{list}, "resource", &p.Resources))
	err = multierror.Append(err, ParseList([]*ast.ObjectList{list}, "provider", &p.Providers))

	err = err.(*multierror.Error).ErrorOrNil()
	return
}

func DefaultPod(namespace string) (p *Pod) {
	p = &Pod{
		Namespace: namespace,
		Target:    defaultPodTarget,
		Runtime:   true,
	}
	return
}

func (p *Pod) Mark() (res uint64) {
	buf, _ := json.Marshal(p)
	res = crc64.Checksum(buf, crc64.MakeTable(crc64.ECMA))
	return
}

/*
Returns Constraint for Resource Request. Returned Constraint includes Pod
and resource request constraints. All manually declared pairs with "resource.*"
and "__*" variables will be excluded. For example Pod which requires "port" and
"counter" resources will return following:

	// Not changed
	* "${meta.test}" = "1"

	// defined manually
	- "${resource.port.pod-one.8080.allocated}" = "true"
	- "${__resource.request.kind.port.allow}" = "true"

	// added by resource kind
	+ "${__resource.request.kind.port.allow}" = "true"
	+ "${__resource.request.kind.counter.allow}" = "true"

	// added to allow requests
	+ "${__resource.request.allow}" = "true"

For Pods without resources "resources.*" and "__.*" will be also excluded. Only
one extra constraint will be added:

	+ "${__resource.request.allow}" = "false"
*/
func (p *Pod) GetResourceRequestConstraint() (res Constraint) {
	res = p.Constraint.FilterOut(openResourcePrefix, hiddenPrefix)
	if len(p.Resources) == 0 {
		res = res.Merge(Constraint{
			fmt.Sprintf("${%s.allow}", resourceRequestPrefix): "false",
		})
		return
	}
	requests := []Constraint{}
	for _, resource := range p.Resources {
		requests = append(requests, resource.GetRequestConstraint())
	}
	res = res.Merge(requests...)
	return
}

// Returns Pod Constraint with additional Resource Allocation constraints. All
// manually defined hidden constraints will be excluded.
func (p *Pod) GetResourceAllocationConstraint() (res Constraint) {
	res = p.Constraint.FilterOut(hiddenPrefix)
	if len(p.Resources) > 0 {
		resourceConstraint := []Constraint{}
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
