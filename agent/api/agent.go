package api

import (
	"github.com/akaspin/soil/agent/api/api-server"
	"github.com/akaspin/soil/proto"
	"net/http"
	"os"
	"syscall"
)

func NewAgentReloadPut(fn func()) (e *api_server.Endpoint) {
	return api_server.NewEndpoint(http.MethodPut, proto.V1AgentReload,
		NewWrapper(func() (err error) {
			fn()
			return
		}))
}

func NewAgentStopPut(signalChan chan os.Signal) (e *api_server.Endpoint) {
	return api_server.NewEndpoint(http.MethodPut, proto.V1AgentStop,
		NewWrapper(func() (err error) {
			defer func() {
				go func() {
					signalChan <- syscall.SIGTERM
				}()
			}()
			return
		}))
}

func NewAgentDrainPut(fn func(on bool)) (e *api_server.Endpoint) {
	e = api_server.NewEndpoint(http.MethodPut, proto.V1AgentDrain, NewWrapper(func() (err error) {
		fn(true)
		return
	}))
	return
}

func NewAgentDrainDelete(fn func(on bool)) (e *api_server.Endpoint) {
	e = api_server.NewEndpoint(http.MethodDelete, proto.V1AgentDrain, NewWrapper(func() (err error) {
		fn(false)
		return
	}))
	return
}
