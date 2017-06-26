package cluster

import (
	"github.com/hashicorp/memberlist"
)

// MergeDelegate Handles node merge
type MergeDelegate struct {

}

func (d *MergeDelegate) NotifyMerge(peers []*memberlist.Node) (err error) {
	return 
}
