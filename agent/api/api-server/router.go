package api_server

import (
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/proto"
	"net/http"
	"net/http/httputil"
	"sync"
)

const (
	defaultHttpScheme  = "http"
	queryParamNode     = "node"
	queryParamRedirect = "redirect"
)

type Router struct {
	log       *logx.Log
	endpoints []*Endpoint
	mux       *http.ServeMux

	nodesMu *sync.RWMutex
	nodes   map[string]string
}

func NewRouter(log *logx.Log, endpoints ...*Endpoint) (r *Router) {
	r = &Router{
		log:       log.GetLog("api", "router"),
		endpoints: endpoints,
		mux:       http.NewServeMux(),
		nodesMu:   &sync.RWMutex{},
		nodes:     map[string]string{},
	}
	paths := map[string][]*Endpoint{}
	for _, endpoint := range endpoints {
		paths[endpoint.path] = append(paths[endpoint.path], endpoint)
	}
	for path, recs := range paths {
		r.mux.HandleFunc(path, r.newHandler(recs))
	}
	r.mux.HandleFunc("/", r.notFoundHandlerFunc)
	return
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.log.Debugf("accepted %s %s", req.Method, req.URL)

	nodeId := req.FormValue(queryParamNode)
	switch nodeId {
	case "", "self":
		r.mux.ServeHTTP(w, req)
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

// ConsumeMessage accepts message with proto.NodesInfo
func (r *Router) ConsumeMessage(message bus.Message) (err error) {
	go func() {
		r.nodesMu.Lock()
		defer r.nodesMu.Unlock()
		var value proto.NodesInfo
		if err := message.Payload().Unmarshal(&value); err != nil {
			r.log.Error(err)
			return
		}
		nodes := map[string]string{}
		for _, node := range value {
			nodes[node.ID] = node.Advertise
		}
		r.nodes = nodes
		r.log.Infof("nodes updated: %v", r.nodes)
	}()
	return
}

func (r *Router) newHandler(endpoints []*Endpoint) (fn func(w http.ResponseWriter, req *http.Request)) {
	get := r.notAllowedHandlerFunc
	put := r.notAllowedHandlerFunc
	del := r.notAllowedHandlerFunc
	for _, endpoint := range endpoints {
		switch endpoint.method {
		case http.MethodGet:
			get = endpoint.getHandleFunc(r.log)
		case http.MethodPut:
			put = endpoint.getHandleFunc(r.log)
		case http.MethodDelete:
			del = endpoint.getHandleFunc(r.log)
		}
	}
	fn = func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			get(w, req)
		case http.MethodPut:
			put(w, req)
		case http.MethodDelete:
			del(w, req)
		default:
			r.notAllowedHandlerFunc(w, req)
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
