// +build ide test_unit

package api_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/api"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httputil"
	"net/url"
	"testing"
	"time"
)

type route1 struct{}

func (r *route1) Empty() interface{} {
	return nil
}

func (*route1) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	res = map[string]string{
		"url": u.String(),
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

	resp, err := http.Get("http://127.0.0.1:3000/v1/route/1")
	assert.NoError(t, err)
	raw, err := httputil.DumpResponse(resp, true)
	assert.NoError(t, err)
	t.Log(string(raw))
}
