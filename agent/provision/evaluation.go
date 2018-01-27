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
	return e
}

func (e *Evaluation) Name() (res string) {
	return e.name
}

func (e *Evaluation) Plan() (res []Instruction) {
	return e.plan
}

func (e *Evaluation) String() (res string) {
	res = e.name + ":"
	if e.Left != nil {
		res += fmt.Sprintf("[%x:%x]", e.Left.PodMark, e.Left.AgentMark)
	} else {
		res += "nil"
	}
	res += "->"
	if e.Right != nil {
		res += fmt.Sprintf("[%x:%x]", e.Right.PodMark, e.Right.AgentMark)
	} else {
		res += "nil"
	}
	return res
}

func (e *Evaluation) Explain() string {
	return fmt.Sprintf("%s", e.plan)
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
		return res
	}

	if e.Left == nil {
		res = append(res, planUnit(nil, e.Right.GetPodUnit())...)
		for _, u := range e.Right.Units {
			res = append(res, planUnit(nil, u)...)
		}
		for _, b := range e.Right.Blobs {
			res = append(res, PlanBlob(nil, b)...)
		}
		return res
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
	return res
}

func planUnit(left, right *allocation.Unit) (res []Instruction) {
	if right == nil {
		return planUnitDestroy(left)
	}

	if left == nil {
		return planUnitDeploy(right.UnitFile, right.Permanent, right.Transition.Create)
	}
	if left.UnitFile.Path != right.UnitFile.Path {
		// unit path changed: generate destroy/create
		res = append(res, planUnitDestroy(left)...)
		res = append(res, planUnitDeploy(right.UnitFile, right.Permanent, right.Transition.Create)...)
		return res
	}
	if left.UnitFile.Source != right.UnitFile.Source {
		return planUnitDeploy(right.UnitFile, right.Permanent, right.Transition.Update)
	}
	// just permanency check
	if left.Permanent != right.Permanent {
		res = append(res, planUnitPerm(right.UnitFile, right.Permanent))
	}

	return res
}

func planUnitDestroy(what *allocation.Unit) (res []Instruction) {
	res = []Instruction{
		NewDeleteUnitInstruction(what.UnitFile),
	}
	if what.Transition.Destroy != "" {
		res = append(res, NewCommandInstruction(phaseDestroyCommand, what.UnitFile, what.Transition.Destroy))
	}
	return res
}

func planUnitDeploy(what allocation.UnitFile, permanent bool, command string) (res []Instruction) {
	res = append(res, NewWriteUnitInstruction(what), planUnitPerm(what, permanent))
	if command != "" {
		res = append(res, NewCommandInstruction(phaseDeployCommand, what, command))
	}
	return res
}

func planUnitPerm(what allocation.UnitFile, permanent bool) (res Instruction) {
	if permanent {
		return NewEnableUnitInstruction(what)
	}
	return NewDisableUnitInstruction(what)
}

func PlanBlob(left, right *allocation.Blob) (res []Instruction) {
	if left == nil && right == nil {
		return nil
	}
	if left == nil {
		return append(res, NewWriteBlobInstruction(phaseDeployFS, right))
	}
	if right == nil {
		if !left.Leave {
			res = append(res, NewDestroyBlobInstruction(phaseDestroyBlobs, left))
		}
		return res
	}
	// ok we have two blobs
	if left.Source != right.Source || left.Permissions != right.Permissions {
		res = append(res, NewWriteBlobInstruction(phaseDeployFS, right))
	}
	return res
}
