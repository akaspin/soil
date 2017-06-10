package allocation

import "fmt"

func Compare(left, right *Pod) (ok bool) {
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

func ToString(p *Pod) (res string) {
	if p == nil {
		res = "<nil>"
		return
	}
	res = fmt.Sprintf("%v", p.Header)
	return
}
