package provision

import (
	"fmt"
	"github.com/akaspin/soil/agent/allocation"
	"sort"
)

type Evaluation struct {
	Left  *allocation.Pod
	Right *allocation.Pod

	name string
	plan []Instruction
}

func NewEvaluation(left, right *allocation.Pod) (e *Evaluation) {
	e = &Evaluation{
		Left:  left,
		Right: right,
		name:  "unknown",
	}
	if right != nil {
		e.name = right.Name
	} else if left != nil {
		e.name = left.Name
	}
	e.plan = e.planPhases()
	sort.Slice(e.plan, func(i, j int) bool {
		return e.plan[i].String() < e.plan[j].String()
	})
	return
}

func (e *Evaluation) Name() (res string) {
	res = e.name
	return
}

func (e *Evaluation) Plan() (res []Instruction) {
	res = e.plan
	return
}

func (e *Evaluation) String() string {
	return fmt.Sprintf("%s:%s", e.name, e.plan)
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
		res = append(res, planUnit(nil, e.Right.GetPodUnit())...)
		for _, u := range e.Right.Units {
			res = append(res, planUnit(nil, u)...)
		}
		for _, b := range e.Right.Blobs {
			res = append(res, PlanBlob(nil, b)...)
		}
		return
	}

	// ok. hard case
	res = append(res, planUnit(e.Left.GetPodUnit(), e.Right.GetPodUnit())...)

	unitsDone := map[string]bool{}
	unitsCandidates := map[string]*allocation.Unit{}

	for _, u := range e.Right.Units {
		unitsCandidates[u.UnitFile.UnitName()] = u
	}
	for _, u := range e.Left.Units {
		res = append(res, planUnit(u, unitsCandidates[u.UnitName()])...)
		unitsDone[u.UnitName()] = true
	}
	for _, u := range e.Right.Units {
		if _, ok := unitsDone[u.UnitFile.UnitName()]; ok {
			continue
		}
		res = append(res, planUnit(nil, u)...)
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

func planUnit(left, right *allocation.Unit) (res []Instruction) {
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
