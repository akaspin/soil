package allocation

import "fmt"

// Returns true if both allocations are is equal or <nil>
func IsEqual(left, right *Pod) (ok bool) {
	var leftMark, rightMark uint64
	if left != nil {
		leftMark = left.Header.Mark()
	}
	if right != nil {
		rightMark = right.Header.Mark()
	}
	ok = leftMark == rightMark
	return
}

// Returns true if right allocation is blocked by left allocation.
func IsBlocked(left, right *Pod) (blocked bool) {
	if left == nil || right == nil {
		return
	}
	leftUnits := map[string]struct{}{}
	for _, unit := range left.Units {
		leftUnits[unit.UnitName()] = struct{}{}
	}
	for _, unit := range right.Units {
		if _, ok := leftUnits[unit.UnitName()]; ok {
			blocked = true
			return
		}
	}

	return
}

func ToString(p *Pod) (res string) {
	if p == nil {
		res = "<nil>"
		return
	}
	res = fmt.Sprintf("%v", p.Header)
	return
}
