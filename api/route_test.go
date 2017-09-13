// +build ide test_unit

package api_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/api"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
	"time"
	"encoding/json"
)

type route1 struct{}

func (r *route1) Empty() interface{} {
	return nil
}

func (*route1) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	res = map[string]interface{}{
		"url": u.Path,
		"params": u.Query(),
	}
	return
}

func TestRouter_Bind(t *testing.T) {

	log := logx.GetLog("test")
	ctx := context.Background()

	router := api.NewRouter()
	router.Get("/v1/route", &route1{})
	router.Get("/v1/route/", &route1{})

	mux := http.NewServeMux()
	router.Bind(ctx, log, mux)
	go http.ListenAndServe(":3000", mux)

	time.Sleep(time.Second)

	resp, err := http.Get("http://127.0.0.1:3000/v1/route/1?param=test")
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	var jsonResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	assert.NoError(t, err)
	assert.Equal(t, jsonResp, map[string]interface{}{
		"url": "/v1/route/1",
		"params": map[string]interface {}{
			"param":[]interface {}{"test"},
		},
	})
}
