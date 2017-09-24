package api

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/akaspin/logx"
	"net/http"
)

type Route struct {
	path     string
	method   string
	endpoint Endpoint
}

// Returns GET route
func GET(path string, endpoint Endpoint) (r *Route) {
	r = newRoute(HttpMethodGET, path, endpoint)
	return
}

// Returns PUT route
func PUT(path string, endpoint Endpoint) (r *Route) {
	r = newRoute(HttpMethodPUT, path, endpoint)
	return
}

// Returns DELETE route
func DELETE(path string, endpoint Endpoint) (r *Route) {
	r = newRoute(HttpMethodDELETE, path, endpoint)
	return
}

func newRoute(method, path string, endpoint Endpoint) (r *Route) {
	r = &Route{
		method:   method,
		path:     path,
		endpoint: endpoint,
	}
	return
}

// generates local HTTP handler
func (r *Route) getHandleFunc(ctx context.Context, log *logx.Log) (h func(w http.ResponseWriter, req *http.Request)) {
	h = func(w http.ResponseWriter, req *http.Request) {
		var err error
		empty := r.endpoint.Empty()
		if empty != nil {
			func() {
				defer req.Body.Close()
				dec := json.NewDecoder(req.Body)
				if err = dec.Decode(&empty); err != nil {
					sendCode(log, w, req, NewError(http.StatusInternalServerError, "can't parse request"))
					return
				}
				req.Body.Close()
			}()
		}
		var data interface{}
		if data, err = r.endpoint.Process(ctx, req.URL, empty); err != nil {
			sendCode(log, w, req, err)
			return
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
			return
		}
		w.Write(raw)
	}
	return
}
