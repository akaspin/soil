package api

import (
	"github.com/akaspin/soil/agent/api/api-server"
)

func NewStatusPingGet() (e *api_server.Endpoint) {
	return api_server.GET("/v1/status/ping", NewWrapper(func() (err error) {
		return
	}))
}
