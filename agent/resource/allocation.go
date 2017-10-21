package resource

import (
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/manifest"
	"github.com/mitchellh/copystructure"
	"github.com/mitchellh/hashstructure"
)

type Alloc struct {
	PodName string
	Request manifest.Resource
	Values  bus.Message
}

func (a Alloc) GetID() string {
	return a.Request.GetID(a.PodName)
}

func (a Alloc) Clone() (res Alloc) {
	res1, _ := copystructure.Copy(a)
	res = res1.(Alloc)
	return
}

func (a Alloc) NeedChange(right manifest.Resource) (res bool) {
	leftHash, _ := hashstructure.Hash(a.Request, nil)
	rightHash, _ := hashstructure.Hash(right, nil)
	res = leftHash != rightHash
	return
}
