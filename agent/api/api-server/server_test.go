// +build ide test_unit

package api_server_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/api/api-server"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/fixture"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

type jsonEndpoint struct {
	id string
}

func (r *jsonEndpoint) Empty() interface{} {
	return nil
}

func (r *jsonEndpoint) Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error) {
	res = map[string]interface{}{
		"id":     r.id,
		"url":    u.Path,
		"params": u.Query(),
	}
	return
}

func checkGetResponse(t *testing.T, uri string, expect map[string]interface{}) {
	t.Helper()
	resp, err := http.Get(uri)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
	var res map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&res)
	assert.Equal(t, expect, res)
}

func TestNewServer(t *testing.T) {
	log := logx.GetLog("test")
	addr := fmt.Sprintf(":%d", fixture.RandomPort(t))
	server := api_server.NewServer(context.Background(), log, addr, api_server.NewRouter(log,
		api_server.NewEndpoint(http.MethodGet, "/v1/route/", &jsonEndpoint{}),
	))
	assert.NoError(t, server.Open())

	time.Sleep(time.Second)

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1%s/v1/route/1", addr))
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	server.Close()
	server.Wait()
}

func TestRouter_ConsumeMessage(t *testing.T) {
	//t.SkipNow()
	log := logx.GetLog("test")
	ctx := context.Background()

	router1 := api_server.NewRouter(log,
		api_server.GET("/v1/route", &jsonEndpoint{"node-1"}),
	)
	router2 := api_server.NewRouter(log,
		api_server.GET("/v1/route", &jsonEndpoint{"node-2"}),
	)

	ts1 := httptest.NewServer(router1)
	defer ts1.Close()
	ts2 := httptest.NewServer(router2)
	defer ts1.Close()

	nodesProducer := bus.NewFlatMap(ctx, log, true, "nodes", router1, router2)
	nodesProducer.Set(map[string]string{
		"node-1": ts1.Listener.Addr().String(),
		"node-2": ts2.Listener.Addr().String(),
	})
	time.Sleep(time.Second)

	checkGetResponse(t,
		ts1.URL+"/v1/route?test=2",
		map[string]interface{}{
			"id":     "node-1",
			"url":    "/v1/route",
			"params": map[string]interface{}{"test": []interface{}{"2"}},
		})
	checkGetResponse(t,
		ts2.URL+"/v1/route?test=2",
		map[string]interface{}{
			"id":     "node-2",
			"url":    "/v1/route",
			"params": map[string]interface{}{"test": []interface{}{"2"}},
		})
	checkGetResponse(t,
		ts1.URL+"/v1/route?node=node-2&redirect&test=2",
		map[string]interface{}{
			"id":     "node-2",
			"url":    "/v1/route",
			"params": map[string]interface{}{"test": []interface{}{"2"}},
		})
	checkGetResponse(t,
		ts1.URL+"/v1/route?node=node-2&test=2",
		map[string]interface{}{
			"id":     "node-2",
			"url":    "/v1/route",
			"params": map[string]interface{}{"test": []interface{}{"2"}},
		})

}
