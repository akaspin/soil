package manifest

import (
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"strings"
)

type Blobs []Blob

func (b *Blobs) Empty() ObjectParser {
	return &Blob{
		Permissions: 0644,
	}
}

func (b *Blobs) Append(v interface{}) (err error) {
	v1 := v.(*Blob)
	*b = append(*b, *v1)
	return
}

// Pod file
type Blob struct {
	Name        string
	Permissions int  `json:",omitempty"`
	Leave       bool `json:",omitempty"`
	Source      string
}

func (b Blob) GetID(parent ...string) string {
	return strings.Join(append(parent, b.Name), ".")
}

func (b *Blob) ParseAST(raw *ast.ObjectItem) (err error) {
	b.Name = raw.Keys[0].Token.Value().(string)
	err = hcl.DecodeObject(b, raw)
	b.Source = Heredoc(b.Source)
	return
}
