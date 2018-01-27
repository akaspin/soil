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
	return leftMark == rightMark
}

// Returns true if right allocation is blocked by left allocation.
func IsBlocked(left, right *Pod) (err error) {
	if left == nil || right == nil {
		return nil
	}
	leftUnits := map[string]struct{}{}
	for _, unit := range left.Units {
		leftUnits[unit.UnitName()] = struct{}{}
	}
	for _, unit := range right.Units {
		name := unit.UnitName()
		if _, ok := leftUnits[name]; ok {
			return fmt.Errorf(`%s blocked by %s(unit:%s)`, left.Name, right.Name, right.Name)
		}
	}
	return nil
}

func ToString(p *Pod) (res string) {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v", p.Header)
}
