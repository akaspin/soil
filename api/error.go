package api

import (
	"fmt"
	"net/http"
)

type Error struct {
	Code   int
	Reason string
}

func NewError(code int, reason string) (e *Error) {
	e = &Error{
		Code:   code,
		Reason: reason,
	}
	return
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d %s", e.Code, e.Reason)
}

func UnwrapError(in error) (out *Error) {
	if in == nil {
		return
	}
	maybe, ok := in.(*Error)
	if ok {
		out = maybe
		return
	}
	out = NewError(http.StatusInternalServerError, in.Error())
	return
}
