package resource

import (
	"github.com/akaspin/soil/agent/allocation"
	"github.com/mitchellh/hashstructure"
)

func Plan(left, right allocation.ResourceSlice) (c, u, d allocation.ResourceSlice) {
	leftByName := map[string]*allocation.Resource{}
	rightByName := map[string]struct{}{}

	for _, l := range left {
		leftByName[l.Request.Name] = l
	}
	for _, r := range right {
		name := r.Request.Name
		rightByName[name] = struct{}{}
		if fromLeft, ok := leftByName[name]; !ok {
			// nothing - add to create
			c = append(c, r)
		} else {
			lh, _ := hashstructure.Hash(fromLeft.Request, nil)
			rh, _ := hashstructure.Hash(r.Request, nil)
			if lh != rh {
				u = append(u, r)
			}
		}
	}
	for _, l := range left {
		if _, ok := rightByName[l.Request.Name]; !ok {
			d = append(d, l)
		}
	}
	return
}
