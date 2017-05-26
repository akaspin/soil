package scheduler

import (
	"sort"
)



func Plan(left, right *Allocation) (res []Instruction) {
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

func planPhases(left, right *Allocation) (res []Instruction) {
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
	candidates := map[string]*AllocationUnit{}
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

func PlanUnit(left, right *AllocationUnit) (res []Instruction) {
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

func planUnitDestroy(what *AllocationUnit) (res []Instruction) {
	res = []Instruction{
		NewDeleteUnitInstruction(what.AllocationFile),
	}
	if what.Transition.Destroy != "" {
		res = append(res, NewCommandInstruction(phaseDestroyCommand, what.AllocationFile, what.Transition.Destroy))
	}
	return
}

func planUnitDeploy(what *AllocationFile, permanent bool, command string) (res []Instruction) {
	res = append(res, NewWriteUnitInstruction(what), planUnitPerm(what, permanent))
	if command != "" {
		res = append(res, NewCommandInstruction(phaseDeployCommand, what, command))
	}
	return
}

func planUnitPerm(what *AllocationFile, permanent bool) (res Instruction) {
	if permanent {
		res = NewEnableUnitInstruction(what)
		return
	}
	res = NewDisableUnitInstruction(what)
	return
}
