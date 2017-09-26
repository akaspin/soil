package api

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/supervisor"
	"net/http"
	"net/http/httputil"
	"sync"
)

const (
	HttpMethodGET    = "GET"
	HttpMethodPUT    = "PUT"
	HttpMethodDELETE = "DELETE"

	queryParamNode     = "node"
	queryParamRedirect = "redirect"

	defaultHttpScheme = "http"
)

type Router struct {
	*supervisor.Control
	log    *logx.Log
	routes []*Route

	nodesMu *sync.RWMutex
	nodes   map[string]string
}

func NewRouter(ctx context.Context, log *logx.Log, routes ...*Route) (r *Router) {
	r = &Router{
		Control: supervisor.NewControl(ctx),
		log:     log.GetLog("api", "router"),
		routes:  routes,
		nodesMu: &sync.RWMutex{},
	}
	return
}

func (r *Router) Bind(ctx context.Context, log *logx.Log, mux *http.ServeMux) {
	// build map
	routes := map[string][]*Route{}
	for _, route := range r.routes {
		routes[route.path] = append(routes[route.path], route)
	}
	for path, recs := range routes {
		mux.HandleFunc(path, r.newHandler(recs))
	}
	mux.HandleFunc("/", r.notFoundHandlerFunc)
}

func (r *Router) ConsumeMessage(message bus.Message) {
	go func() {
		r.nodesMu.Lock()
		defer r.nodesMu.Unlock()
		r.nodes = message.GetPayload()
		r.log.Debugf("synced nodes: %v", message.GetPayload())
	}()
}

func (r *Router) newHandler(records []*Route) (fn func(w http.ResponseWriter, req *http.Request)) {
	get := r.notAllowedHandlerFunc
	put := r.notAllowedHandlerFunc
	del := r.notAllowedHandlerFunc
	for _, record := range records {
		switch record.method {
		case HttpMethodGET:
			get = record.getHandleFunc(r.Control.Ctx(), r.log)
		case HttpMethodPUT:
			put = record.getHandleFunc(r.Control.Ctx(), r.log)
		case HttpMethodDELETE:
			del = record.getHandleFunc(r.Control.Ctx(), r.log)
		}
	}
	fn = func(w http.ResponseWriter, req *http.Request) {
		r.log.Tracef("accepted %s %s", req.Method, req.URL)

		nodeId := req.FormValue(queryParamNode)
		switch nodeId {
		case "", "self":
			switch req.Method {
			case HttpMethodGET:
				get(w, req)
			case HttpMethodPUT:
				put(w, req)
			case HttpMethodDELETE:
				del(w, req)
			default:
				r.notAllowedHandlerFunc(w, req)
			}
		default:
			r.nodesMu.RLock()
			nodeAddr, ok := r.nodes[nodeId]
			r.nodesMu.RUnlock()
			if !ok {
				sendCode(r.log, w, req, NewError(404, "node not found"))
				return
			}
			_, canRedirect := req.URL.Query()[queryParamRedirect]
			targetUrl, err := req.URL.Parse(fmt.Sprintf("%s://%s", defaultHttpScheme, nodeAddr))
			if err != nil {
				sendCode(r.log, w, req, err)
				return
			}

			// remove node and redirect from query
			values := req.URL.Query()
			values.Del(queryParamNode)
			values.Del(queryParamRedirect)
			req.URL.RawQuery = values.Encode()

			// check client permits redirects
			if canRedirect {
				r.log.Debugf("redirecting %s %s to %s (%s)", req.Method, req.URL, nodeId, nodeAddr)
				targetUrl = targetUrl.ResolveReference(req.URL)
				sendCode(r.log, w, req, NewError(http.StatusMovedPermanently, targetUrl.String()))
				return
			}

			// proxy if can't redirect
			r.log.Debugf("proxying %s %s to %s (%s)", req.Method, req.URL, nodeId, nodeAddr)
			proxy := httputil.NewSingleHostReverseProxy(targetUrl)
			proxy.ServeHTTP(w, req)
		}
	}
	return
}

func (r *Router) notAllowedHandlerFunc(w http.ResponseWriter, req *http.Request) {
	sendCode(r.log, w, req, errorMethodsNotAllowed)
}

func (r *Router) notFoundHandlerFunc(w http.ResponseWriter, req *http.Request) {
	sendCode(r.log, w, req, NewError(http.StatusNotFound, fmt.Sprintf("not found %s", req.URL)))
}
