package api_server

import (
	"fmt"
	"github.com/akaspin/logx"
	"net/http"
)

var (
	errorMethodsNotAllowed = NewError(http.StatusMethodNotAllowed, "method is not allowed")
	ErrorBadRequestData    = NewError(http.StatusBadRequest, "bad request data")
)

type Error struct {
	Code   int
	Reason string
}

func NewError(code int, reason string) (e *Error) {
	return &Error{
		Code:   code,
		Reason: reason,
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d %s", e.Code, e.Reason)
}

func UnwrapError(in error) (out *Error) {
	if in == nil {
		return nil
	}
	maybe, ok := in.(*Error)
	if ok {
		return maybe
	}
	return NewError(http.StatusInternalServerError, in.Error())
}

func sendCode(log *logx.Log, w http.ResponseWriter, req *http.Request, handlerErr error) {
	parsed := NewError(http.StatusInternalServerError, "unknown")
	if unwrapped := UnwrapError(handlerErr); unwrapped != nil {
		parsed = unwrapped
	}
	if parsed.Code >= http.StatusBadRequest {
		log.Errorf("%s %s %s", req.Method, req.URL.String(), parsed.Error())
	}
	switch parsed.Code {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther:
		http.Redirect(w, req, parsed.Reason, parsed.Code)
	default:
		http.Error(w, parsed.Reason, parsed.Code)
	}
}
