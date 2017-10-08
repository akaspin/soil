package resource

import (
	"github.com/akaspin/soil/agent/allocation"
	"github.com/mitchellh/copystructure"
)

// Resource allocation
type Allocation struct {
	PodName string
	*allocation.Resource
}

func (a *Allocation) GetId() string {
	return a.Request.GetID(a.PodName)
}

func (a *Allocation) Clone() (res *Allocation) {
	res1, _ := copystructure.Copy(a)
	res = res1.(*Allocation)
	return
}