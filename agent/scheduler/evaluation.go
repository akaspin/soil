package scheduler

import (
	"github.com/akaspin/soil/agent/allocation"
	"sort"
)

type Evaluation struct {
	Left  *allocation.Pod
	Right *allocation.Pod
}

func (e *Evaluation) Name() (res string) {
	if e.Right != nil {
		res = e.Right.Name
		return
	}
	if e.Left != nil {
		res = e.Left.Name
	}
	return
}

func (e *Evaluation) Plan() (res []Instruction) {
	phases1 := map[int][]Instruction{}
	var phaseIds []int

	for _, i := range e.planPhases() {
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

func (e *Evaluation) planPhases() (res []Instruction) {
	if e.Right == nil {
		res = append(res, planUnitDestroy(e.Left.GetPodUnit())...)
		for _, u := range e.Left.Units {
			res = append(res, planUnitDestroy(u)...)
		}
		for _, b := range e.Left.Blobs {
			res = append(res, PlanBlob(b, nil)...)
		}
		return
	}

	if e.Left == nil {
		res = append(res, PlanUnit(nil, e.Right.GetPodUnit())...)
		for _, u := range e.Right.Units {
			res = append(res, PlanUnit(nil, u)...)
		}
		for _, b := range e.Right.Blobs {
			res = append(res, PlanBlob(nil, b)...)
		}
		return
	}

	// ok. hard case
	res = append(res, PlanUnit(e.Left.GetPodUnit(), e.Right.GetPodUnit())...)

	unitsDone := map[string]bool{}
	unitsCandidates := map[string]*allocation.Unit{}

	for _, u := range e.Right.Units {
		unitsCandidates[u.UnitFile.UnitName()] = u
	}
	for _, u := range e.Left.Units {
		res = append(res, PlanUnit(u, unitsCandidates[u.UnitName()])...)
		unitsDone[u.UnitName()] = true
	}
	for _, u := range e.Right.Units {
		if _, ok := unitsDone[u.UnitFile.UnitName()]; ok {
			continue
		}
		res = append(res, PlanUnit(nil, u)...)
	}

	blobsDone := map[string]bool{}
	blobCandidates := map[string]*allocation.Blob{}
	for _, b := range e.Right.Blobs {
		blobCandidates[b.Name] = b
	}
	for _, b := range e.Left.Blobs {
		res = append(res, PlanBlob(b, blobCandidates[b.Name])...)
		blobsDone[b.Name] = true
	}
	for _, b := range e.Right.Blobs {
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
			res = append(res, NewDestroyBlobInstruction(phaseDestroyBlobs, left))
		}
		return
	}
	// ok we have two blobs
	if left.Source != right.Source || left.Permissions != right.Permissions {
		res = append(res, NewWriteBlobInstruction(phaseDeployFS, right))
	}
	return
}
