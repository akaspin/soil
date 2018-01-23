package errslice

// Error is simple slice of errors
type Error []error

// Error returns all collected errors delimited by comma
func (e Error) Error() (res string) {
	for i, e1 := range e {
		if i > 0 {
			res += ","
		}
		res += e1.Error()
	}
	return
}

