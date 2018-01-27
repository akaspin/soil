package lib

import (
	"bytes"
	"fmt"
	"github.com/akaspin/errslice"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"io"
)

func ParseHCLMerge(readers ...io.Reader) (roots *ast.ObjectList, err error) {
	var err1 error
	contents := &bytes.Buffer{}
	var ok bool
	for _, reader := range readers {
		var buf bytes.Buffer
		if _, err1 = io.Copy(&buf, reader); err1 != nil {
			err = errslice.Append(err, err1)
			continue
		}
		var rawRoot *ast.File
		if rawRoot, err1 = hcl.Parse(buf.String()); err1 != nil {
			err = errslice.Append(err, err1)
			continue
		}
		_, ok := rawRoot.Node.(*ast.ObjectList)
		if !ok {
			err = errslice.Append(err, fmt.Errorf("error parsing: root should be an object"))
			continue
		}
		contents.Write(buf.Bytes())
		contents.WriteString("\n")
	}
	var rawRoot *ast.File
	if rawRoot, err1 = hcl.Parse(contents.String()); err1 != nil {
		return nil, errslice.Append(err, err1)
	}
	roots, ok = rawRoot.Node.(*ast.ObjectList)
	if !ok {
		err = errslice.Append(err, fmt.Errorf("error parsing: root should be an object"))
	}
	return roots, err
}

func ParseHCL(readers ...io.Reader) (lists []*ast.ObjectList, err error) {
	var err1 error
	var ok bool
	for _, reader := range readers {
		var buf bytes.Buffer
		if _, err1 = io.Copy(&buf, reader); err1 != nil {
			err = errslice.Append(err, err1)
			continue
		}
		var rawRoot *ast.File
		if rawRoot, err1 = hcl.Parse(buf.String()); err1 != nil {
			err = errslice.Append(err, err1)
			continue
		}
		var list *ast.ObjectList
		list, ok = rawRoot.Node.(*ast.ObjectList)
		if !ok {
			err = errslice.Append(err, fmt.Errorf("error parsing: root should be an object"))
			continue
		}
		lists = append(lists, list)
	}
	return lists, err
}
