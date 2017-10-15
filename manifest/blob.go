package manifest

import (
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
)

// Pod file
type Blob struct {
	Name        string
	Permissions int
	Leave       bool
	Source      string
}

func defaultBlob() (b Blob) {
	b = Blob{
		Permissions: 0644,
	}
	return
}

func (b *Blob) parseAst(raw *ast.ObjectItem) (err error) {
	b.Name = raw.Keys[0].Token.Value().(string)
	err = hcl.DecodeObject(b, raw)
	b.Source = Heredoc(b.Source)
	return
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
