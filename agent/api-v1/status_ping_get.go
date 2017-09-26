package api_v1

import (
	"github.com/akaspin/soil/agent/api-v1/api-server"
)

func NewStatusPingGet() (e *api_server.Endpoint) {
	return api_server.GET("/v1/status/ping", NewWrapper(func() (err error) {
		return
	}))
}
