package api

import (
	"context"
	"encoding/json"
	"github.com/akaspin/logx"
	"net/http"
)

type Router struct {
	routes []*routeRecord
}

func NewRouter() (r *Router) {
	r = &Router{}
	return
}

func (r *Router) Get(path string, endpoint Endpoint) {
	r.add("GET", path, endpoint)
}

func (r *Router) Put(path string, endpoint Endpoint) {
	r.add("PUT", path, endpoint)
}

func (r *Router) Delete(path string, endpoint Endpoint) {
	r.add("DELETE", path, endpoint)
}

func (r *Router) Bind(ctx context.Context, log *logx.Log, mux *http.ServeMux) {
	// build map
	routes := map[string][]*routeRecord{}
	for _, route := range r.routes {
		routes[route.path] = append(routes[route.path], route)
	}
	na := NewCodeHandlerFunc(log, NewError(http.StatusMethodNotAllowed, "method is not allowed"))
	for path, recs := range routes {
		mux.HandleFunc(path, newMethodHandleFunc(ctx, log, na, recs))
	}
}

func (r *Router) add(method, path string, endpoint Endpoint) {
	r.routes = append(r.routes, &routeRecord{
		path:     path,
		method:   method,
		endpoint: endpoint,
	})
}

func newMethodHandleFunc(ctx context.Context, log *logx.Log, na func(w http.ResponseWriter, req *http.Request), records []*routeRecord) (fn func(w http.ResponseWriter, req *http.Request)) {
	get := na
	put := na
	del := na
	for _, record := range records {
		switch record.method {
		case "GET":
			get = record.getHandleFunc(ctx, log)
		case "PUT":
			put = record.getHandleFunc(ctx, log)
		case "DELETE":
			del = record.getHandleFunc(ctx, log)
		}
	}
	fn = func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case "GET":
			get(w, req)
		case "PUT":
			put(w, req)
		case "DELETE":
			del(w, req)
		default:
			na(w, req)
		}
	}
	return
}

type routeRecord struct {
	path     string
	method   string
	endpoint Endpoint
}

func (r *routeRecord) getHandleFunc(ctx context.Context, log *logx.Log) (h func(w http.ResponseWriter, req *http.Request)) {
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
		w.Write(raw)
	}
	return
}
