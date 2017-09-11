// +build ide test_unit

package api_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/api"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {

	log := logx.GetLog("test")
	ctx := context.Background()

	router := api.NewRouter()
	router.Get("/v1/route", &route1{})
	router.Get("/v1/route/", &route1{})

	server := api.NewServer(ctx, log, ":3333", router)
	assert.NoError(t, server.Open())

	time.Sleep(time.Second)

	resp, err := http.Get("http://127.0.0.1:3333/v1/route/1")
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	server.Close()
	server.Wait()
}
