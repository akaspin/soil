package executor

import (
	"github.com/akaspin/soil/agent/scheduler/allocation"
)

const (
	phaseDestroyCommand = iota
	phaseDestroyFS
	phaseDeployFS
	phaseDeployPerm
	phaseDeployCommand
)

func Plan(left, right *allocation.Allocation) (res []Instruction) {
	phases := map[int][]Instruction{
		phaseDestroyCommand: nil,
		phaseDestroyFS:      nil,
		phaseDeployFS:       nil,
		phaseDeployPerm:     nil,
		phaseDeployCommand:  nil,
	}
	for _, i := range PlanPhases(left, right) {
		phase := i.Phase()
		phases[phase] = append(phases[phase], i)
	}
	for _, phase := range []int{
		phaseDestroyCommand,
		phaseDestroyFS,
		phaseDeployFS,
		phaseDeployPerm,
		phaseDeployCommand,
	} {
		res = append(res, phases[phase]...)
	}
	return
}

func PlanPhases(left, right *allocation.Allocation) (res []Instruction) {
	if right == nil {
		res = append(res, planUnitDestroy(left.PodUnit())...)
		for _, u := range left.Units {
			res = append(res, planUnitDestroy(u)...)
		}
		return
	}

	if left == nil {
		res = append(res, PlanUnit(nil, right.PodUnit())...)
		for _, u := range right.Units {
			res = append(res, PlanUnit(nil, u)...)
		}
		return
	}

	// ok. hard case
	res = append(res, PlanUnit(left.PodUnit(), right.PodUnit())...)

	done := map[string]bool{}
	candidates := map[string]*allocation.AllocationUnit{}
	for _, u := range right.Units {
		candidates[u.AllocationFile.UnitName()] = u
	}

	for _, u := range left.Units {
		res = append(res, PlanUnit(u, candidates[u.UnitName()])...)
		done[u.UnitName()] = true
	}
	for _, u := range right.Units {
		if _, ok := done[u.AllocationFile.UnitName()]; ok {
			continue
		}
		res = append(res, PlanUnit(nil, u)...)
	}

	return
}

func PlanUnit(left, right *allocation.AllocationUnit) (res []Instruction) {
	if right == nil {
		res = planUnitDestroy(left)
		return
	}

	if left == nil {
		res = planUnitDeploy(right.AllocationFile, right.Permanent, right.Transition.Create)
		return
	}
	if left.AllocationFile.Path != right.AllocationFile.Path {
		// unit path changed: generate destroy/create
		res = append(res, planUnitDestroy(left)...)
		res = append(res, planUnitDeploy(right.AllocationFile, right.Permanent, right.Transition.Create)...)
		return
	}
	if left.AllocationFile.Source != right.AllocationFile.Source {
		res = planUnitDeploy(right.AllocationFile, right.Permanent, right.Transition.Update)
		return
	}
	// just permanency check
	if left.Permanent != right.Permanent {
		res = append(res, planUnitPerm(right.AllocationFile, right.Permanent))
	}

	return
}

func planUnitDestroy(what *allocation.AllocationUnit) (res []Instruction) {
	res = []Instruction{
		NewDeleteUnitInstruction(what.AllocationFile),
	}
	if what.Transition.Destroy != "" {
		res = append(res, NewCommandInstruction(phaseDestroyCommand, what.AllocationFile, what.Transition.Destroy))
	}
	return
}

func planUnitDeploy(what *allocation.AllocationFile, permanent bool, command string) (res []Instruction) {
	res = append(res, NewWriteUnitInstruction(what), planUnitPerm(what, permanent))
	if command != "" {
		res = append(res, NewCommandInstruction(phaseDeployCommand, what, command))
	}
	return
}

func planUnitPerm(what *allocation.AllocationFile, permanent bool) (res Instruction) {
	if permanent {
		res = NewEnableUnitInstruction(what)
		return
	}
	res = NewDisableUnitInstruction(what)
	return
}
