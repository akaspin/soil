package manifest

import (
	"encoding/json"
	"github.com/akaspin/errslice"
	"github.com/akaspin/soil/lib"
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

type PodSlice []*Pod

func (r *PodSlice) Empty() ObjectParser {
	return &Pod{
		Namespace: PrivateNamespace,
		Target:    defaultPodTarget,
		Runtime:   true,
	}
}

func (r *PodSlice) Append(v interface{}) (err error) {
	*r = append(*r, v.(*Pod))
	return nil
}

func (r *PodSlice) SetNamespace(namespace string) {
	for _, pod := range *r {
		pod.Namespace = namespace
	}
}

func (r *PodSlice) Unmarshal(namespace string, reader ...io.Reader) (err error) {
	roots, parseErr := lib.ParseHCL(reader...)
	err = errslice.Append(err, parseErr)
	err = errslice.Append(err, ParseList(roots, "pod", r))
	r.SetNamespace(namespace)
	return err
}

// Pod manifest
type Pod struct {
	Namespace  string
	Name       string
	Runtime    bool
	Target     string
	Constraint Constraint    `json:",omitempty"`
	Units      Units         `json:",omitempty" hcl:"-"`
	Blobs      Blobs         `json:",omitempty" hcl:"-"`
	Resources  Resources     `json:",omitempty" hcl:"-"`
	Providers  ProviderSlice `json:",omitempty" hcl:"-"`
}

func (p Pod) GetID(parent ...string) string {
	return strings.Join(append(parent, p.Namespace, p.Name), ".")
}

func (p *Pod) ParseAST(raw *ast.ObjectItem) (err error) {
	list := raw.Val.(*ast.ObjectType).List

	if err = errslice.Append(err, hcl.DecodeObject(p, raw)); err != nil {
		return err
	}
	p.Name = raw.Keys[0].Token.Value().(string)

	err = errslice.Append(err, ParseList([]*ast.ObjectList{list}, "unit", &p.Units))
	err = errslice.Append(err, ParseList([]*ast.ObjectList{list}, "blob", &p.Blobs))
	err = errslice.Append(err, ParseList([]*ast.ObjectList{list}, "resource", &p.Resources))
	err = errslice.Append(err, ParseList([]*ast.ObjectList{list}, "provider", &p.Providers))

	return err
}

// Get Pod checksum
func (p *Pod) Mark() (res uint64) {
	buf, _ := json.Marshal(p)
	return crc64.Checksum(buf, crc64.MakeTable(crc64.ECMA))
}

// Compare
func IsEqual(left, right *Pod) (ok bool) {
	if left == nil {
		if right != nil {
			return false
		}
		return true
	}
	if left.Mark() == right.Mark() {
		return true
	}
	return false
}
