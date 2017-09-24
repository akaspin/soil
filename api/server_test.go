// +build ide test_unit

package api_test

import (
	"context"
	"encoding/json"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/metadata"
	"github.com/akaspin/soil/api"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"net/http"
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

func TestServer_Single(t *testing.T) {
	t.SkipNow()
	log := logx.GetLog("test")
	ctx := context.Background()

	router := api.NewRouter(ctx, log,
		api.GET("/v1/route", &jsonEndpoint{}),
		api.GET("/v1/route/", &jsonEndpoint{}),
	)
	server := api.NewServer(ctx, log, ":3333", router)
	sv := supervisor.NewChain(ctx, router, server)
	assert.NoError(t, sv.Open())

	time.Sleep(time.Second)

	resp, err := http.Get("http://127.0.0.1:3333/v1/route/1")
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	sv.Close()
	sv.Wait()
}

func TestRouter_Sync(t *testing.T) {
	//t.SkipNow()
	// run two servers
	log := logx.GetLog("test")
	ctx := context.Background()

	router1 := api.NewRouter(ctx, log,
		api.GET("/v1/route", &jsonEndpoint{"node-1"}),
	)
	server1 := api.NewServer(ctx, log, ":4444", router1)
	sv1 := supervisor.NewChain(ctx, router1, server1)

	router2 := api.NewRouter(ctx, log,
		api.GET("/v1/route", &jsonEndpoint{"node-2"}),
	)
	server2 := api.NewServer(ctx, log, ":5555", router2)
	sv2 := supervisor.NewChain(ctx, router2, server2)

	nodesProducer := metadata.NewSimpleProducer(ctx, log, "nodes", router1, router2)

	sv := supervisor.NewChain(ctx,
		supervisor.NewGroup(ctx, sv1, sv2),
		nodesProducer,
	)
	assert.NoError(t, sv.Open())

	time.Sleep(time.Second)

	nodesProducer.Replace(map[string]string{
		"node-1": "127.0.0.1:4444",
		"node-2": "127.0.0.1:5555",
	})
	time.Sleep(time.Second)

	checkGetResponse(t,
		"http://127.0.0.1:4444/v1/route?test=2",
		map[string]interface{}{
			"id":     "node-1",
			"url":    "/v1/route",
			"params": map[string]interface{}{"test": []interface{}{"2"}},
		})
	checkGetResponse(t, "http://127.0.0.1:5555/v1/route?test=2",
		map[string]interface{}{
			"id":     "node-2",
			"url":    "/v1/route",
			"params": map[string]interface{}{"test": []interface{}{"2"}},
		})
	checkGetResponse(t, "http://127.0.0.1:4444/v1/route?node=node-2&redirect&test=2",
		map[string]interface{}{
			"id":     "node-2",
			"url":    "/v1/route",
			"params": map[string]interface{}{"test": []interface{}{"2"}},
		})
	checkGetResponse(t, "http://127.0.0.1:4444/v1/route?node=node-2&test=2",
		map[string]interface{}{
			"id":     "node-2",
			"url":    "/v1/route",
			"params": map[string]interface{}{"test": []interface{}{"2"}},
		})

	sv.Close()
	sv.Wait()
}
