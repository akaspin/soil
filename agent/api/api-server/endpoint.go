package api_server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/akaspin/logx"
	"net/http"
)

type Endpoint struct {
	path      string
	method    string
	processor Processor
}

// Returns GET route
func GET(path string, processor Processor) (r *Endpoint) {
	return NewEndpoint(http.MethodGet, path, processor)
}

// Returns PUT route
func PUT(path string, processor Processor) (r *Endpoint) {
	return NewEndpoint(http.MethodPut, path, processor)
}

// Returns DELETE route
func DELETE(path string, processor Processor) (r *Endpoint) {
	return NewEndpoint(http.MethodDelete, path, processor)
}

func NewEndpoint(method, path string, endpoint Processor) (r *Endpoint) {
	return &Endpoint{
		method:    method,
		path:      path,
		processor: endpoint,
	}
}

func (e *Endpoint) Processor() (p Processor) {
	return e.processor
}

// generates local HTTP handler
func (e *Endpoint) getHandleFunc(log *logx.Log) (h func(w http.ResponseWriter, req *http.Request)) {
	return func(w http.ResponseWriter, req *http.Request) {
		var err error
		empty := e.processor.Empty()
		if empty != nil {
			func() {
				defer req.Body.Close()
				dec := json.NewDecoder(req.Body)
				if err = dec.Decode(&empty); err != nil {
					sendCode(log, w, req, NewError(http.StatusInternalServerError, "can't parse request"))
					return //
				}
				req.Body.Close()
			}()
		}
		var data interface{}
		if data, err = e.processor.Process(req.Context(), req.URL, empty); err != nil {
			if err == ErrorBadRequestData {
				sendCode(log, w, req, NewError(http.StatusBadRequest, fmt.Sprintf("bad data (%T)%#v", empty, empty)))
				return //
			}
			sendCode(log, w, req, err)
			return //
		}
		var raw []byte
		if raw, err = json.Marshal(&data); err != nil {
			sendCode(log, w, req, NewError(http.StatusInternalServerError, "can't marshal response"))
		}
		if _, ok := req.URL.Query()["pretty"]; ok {
			// pretty
			var buf bytes.Buffer
			if err = json.Indent(&buf, raw, "", "  "); err != nil {
				sendCode(log, w, req, NewError(http.StatusInternalServerError, "can't marshal response"))
			}
			w.Write(append(buf.Bytes(), "\n"...))
			return //
		}
		w.Write(raw)
		log.Debugf(`ok %s %s: %v`, req.Method, req.URL.String(), data)
	}
}
