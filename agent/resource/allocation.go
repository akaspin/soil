package resource

import (
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/manifest"
	"github.com/mitchellh/copystructure"
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
