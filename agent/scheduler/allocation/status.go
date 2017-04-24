package allocation

import "fmt"

func ComparePods(left, right *Pod) (ok bool) {
	var leftMark, rightMark uint64
	if left != nil {
		leftMark = left.PodHeader.Mark()
	}
	if right != nil {
		rightMark = right.PodHeader.Mark()
	}
	ok = leftMark == rightMark
	return
}

func PodToString(p *Pod) (res string) {
	if p == nil {
		res = "<nil>"
		return
	}
	res = fmt.Sprintf("%v", p.PodHeader)
	return
}

