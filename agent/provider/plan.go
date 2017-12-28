package provider

import (
	"github.com/akaspin/soil/agent/allocation"
	"github.com/mitchellh/hashstructure"
)

// Plan determines changes between left and right for one pod
func Plan(left, right allocation.ProviderSlice) (create, update allocation.ProviderSlice, destroy []string) {
	leftByName := map[string]*allocation.Provider{}
	rightByName := map[string]struct{}{}
	for _, prov := range left {
		leftByName[prov.Name] = prov
	}
	for _, prov := range right {
		name := prov.Name
		rightByName[name] = struct{}{}
		if fromLeft, ok := leftByName[name]; !ok {
			// nothing - add to create
			create = append(create, prov)
		} else {
			lh, _ := hashstructure.Hash(fromLeft, nil)
			rh, _ := hashstructure.Hash(prov, nil)
			if lh != rh {
				update = append(update, prov)
			}
		}
	}
	for _, prov := range left {
		if _, ok := rightByName[prov.Name]; !ok {
			destroy = append(destroy, prov.Name)
		}
	}
	return
}
