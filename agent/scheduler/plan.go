package scheduler

import (
	"github.com/akaspin/soil/agent/allocation"
	"sort"
)

func Plan(left, right *allocation.Pod) (res []Instruction) {
	phases1 := map[int][]Instruction{}
	var phaseIds []int
	for _, i := range planPhases(left, right) {
		phase := i.Phase()
		phases1[phase] = append(phases1[phase], i)
	}
	for i := range phases1 {
		phaseIds = append(phaseIds, i)
	}
	sort.Ints(phaseIds)
	for _, i := range phaseIds {
		res = append(res, phases1[i]...)
	}
	return
}

func planPhases(left, right *allocation.Pod) (res []Instruction) {
	if right == nil {
		res = append(res, planUnitDestroy(left.GetPodUnit())...)
		for _, u := range left.Units {
			res = append(res, planUnitDestroy(u)...)
		}
		for _, b := range left.Blobs {
			res = append(res, PlanBlob(b, nil)...)
		}
		return
	}

	if left == nil {
		res = append(res, PlanUnit(nil, right.GetPodUnit())...)
		for _, u := range right.Units {
			res = append(res, PlanUnit(nil, u)...)
		}
		for _, b := range right.Blobs {
			res = append(res, PlanBlob(nil, b)...)
		}
		return
	}

	// ok. hard case
	res = append(res, PlanUnit(left.GetPodUnit(), right.GetPodUnit())...)

	unitsDone := map[string]bool{}
	unitsCandidates := map[string]*allocation.Unit{}

	for _, u := range right.Units {
		unitsCandidates[u.UnitFile.UnitName()] = u
	}
	for _, u := range left.Units {
		res = append(res, PlanUnit(u, unitsCandidates[u.UnitName()])...)
		unitsDone[u.UnitName()] = true
	}
	for _, u := range right.Units {
		if _, ok := unitsDone[u.UnitFile.UnitName()]; ok {
			continue
		}
		res = append(res, PlanUnit(nil, u)...)
	}

	blobsDone := map[string]bool{}
	blobCandidates := map[string]*allocation.Blob{}
	for _, b := range right.Blobs {
		blobCandidates[b.Name] = b
	}
	for _, b := range left.Blobs {
		res = append(res, PlanBlob(b, blobCandidates[b.Name])...)
		blobsDone[b.Name] = true
	}
	for _, b := range right.Blobs {
		if _, ok := blobsDone[b.Name]; ok {
			continue
		}
		res = append(res, PlanBlob(nil, b)...)
	}

	return
}

func PlanUnit(left, right *allocation.Unit) (res []Instruction) {
	if right == nil {
		res = planUnitDestroy(left)
		return
	}

	if left == nil {
		res = planUnitDeploy(right.UnitFile, right.Permanent, right.Transition.Create)
		return
	}
	if left.UnitFile.Path != right.UnitFile.Path {
		// unit path changed: generate destroy/create
		res = append(res, planUnitDestroy(left)...)
		res = append(res, planUnitDeploy(right.UnitFile, right.Permanent, right.Transition.Create)...)
		return
	}
	if left.UnitFile.Source != right.UnitFile.Source {
		res = planUnitDeploy(right.UnitFile, right.Permanent, right.Transition.Update)
		return
	}
	// just permanency check
	if left.Permanent != right.Permanent {
		res = append(res, planUnitPerm(right.UnitFile, right.Permanent))
	}

	return
}

func planUnitDestroy(what *allocation.Unit) (res []Instruction) {
	res = []Instruction{
		NewDeleteUnitInstruction(what.UnitFile),
	}
	if what.Transition.Destroy != "" {
		res = append(res, NewCommandInstruction(phaseDestroyCommand, what.UnitFile, what.Transition.Destroy))
	}
	return
}

func planUnitDeploy(what *allocation.UnitFile, permanent bool, command string) (res []Instruction) {
	res = append(res, NewWriteUnitInstruction(what), planUnitPerm(what, permanent))
	if command != "" {
		res = append(res, NewCommandInstruction(phaseDeployCommand, what, command))
	}
	return
}

func planUnitPerm(what *allocation.UnitFile, permanent bool) (res Instruction) {
	if permanent {
		res = NewEnableUnitInstruction(what)
		return
	}
	res = NewDisableUnitInstruction(what)
	return
}

func PlanBlob(left, right *allocation.Blob) (res []Instruction) {
	if left == nil && right == nil {
		return
	}
	if left == nil {
		res = append(res, NewWriteBlobInstruction(phaseDeployFS, right))
		return
	}
	if right == nil {
		if !left.Leave {
			res = append(res, NewDestroyBlobInstruction(phaseDeployFS, left))
		}
		return
	}
	// ok we have two blobs
	if left.Source != right.Source || left.Permissions != right.Permissions {
		res = append(res, NewWriteBlobInstruction(phaseDeployFS, right))
	}
	return
}
