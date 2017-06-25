package api

import (
	"github.com/akaspin/logx"
	"net/http"
)

func NewCodeHandlerFunc(log *logx.Log, err error) (f func(w http.ResponseWriter, req *http.Request)) {
	f = func(w http.ResponseWriter, req *http.Request) {
		sendCode(log, w, req, err)
	}
	return
}

func sendCode(log *logx.Log, w http.ResponseWriter, req *http.Request, handlerErr error) (err error) {
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
	return
}
