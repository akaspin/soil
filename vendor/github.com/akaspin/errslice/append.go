package errslice

// Append takes two errors and returns minimal result. If first argument is
// nil Append will return second and vise versa. If both arguments are nil
// Append will return nil.
//
// If both arguments are not nil Append will combine them in Error
func Append(left, right error) (err error) {
	if right == nil {
		return left
	}
	if left == nil {
		return right
	}
	var err1 Error
	if l, ok := left.(Error); ok {
		err1 = append(err1, l...)
	} else {
		err1 = append(err1, left)
	}
	if r, ok := right.(Error); ok {
		err1 = append(err1, r...)
	} else {
		err1 = append(err1, right)
	}
	return err1
}


