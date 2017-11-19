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
	"github.com/akaspin/soil/proto"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
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
	log := logx.GetLog("test")

	router1 := api_server.NewRouter(log,
		api_server.GET("/v1/route", &jsonEndpoint{"node-1"}),
	)
	router2 := api_server.NewRouter(log,
		api_server.GET("/v1/route", &jsonEndpoint{"node-2"}),
	)
	nodesProducer := bus.NewTeePipe(router1, router2)

	ts1 := httptest.NewServer(router1)
	defer ts1.Close()
	ts2 := httptest.NewServer(router2)
	defer ts1.Close()

	checkGet := func(t *testing.T, uri string, code int, expect map[string]interface{}) {
		t.Helper()
		fixture.WaitNoError(t, fixture.DefaultWaitConfig(), func() (err error) {
			resp, err := http.Get(uri)
			if err != nil {
				return
			}
			if resp.StatusCode != code {
				err = fmt.Errorf(`bad status code: %d != %d`, code, resp.StatusCode)
				return
			}
			if expect == nil {
				return
			}
			var res map[string]interface{}
			if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
				return
			}
			if !reflect.DeepEqual(expect, res) {
				err = fmt.Errorf("not equal (expected)%s != (actual)%s", expect, res)
			}
			return
		})
	}

	t.Run(`no nodes node-1 self`, func(t *testing.T) {
		checkGet(t,
			ts1.URL+"/v1/route?test=2",
			200,
			map[string]interface{}{
				"id":     "node-1",
				"url":    "/v1/route",
				"params": map[string]interface{}{"test": []interface{}{"2"}},
			},
		)
	})
	t.Run(`no nodes node-1 proxy to node-2`, func(t *testing.T) {
		checkGet(t,
			ts1.URL+"/v1/route?node=node-2&test=2",
			404,
			nil,
		)
	})
	t.Run(`configure node-1 and node-2`, func(t *testing.T) {
		nodesProducer.ConsumeMessage(bus.NewMessage("nodes", []interface{}{
			proto.NodeInfo{
				ID:        "node-1",
				Advertise: ts1.Listener.Addr().String(),
			},
			proto.NodeInfo{
				ID:        "node-2",
				Advertise: ts2.Listener.Addr().String(),
			},
		}))
	})
	t.Run(`with nodes node-1 self`, func(t *testing.T) {
		checkGet(t,
			ts1.URL+"/v1/route?test=2",
			200,
			map[string]interface{}{
				"id":     "node-1",
				"url":    "/v1/route",
				"params": map[string]interface{}{"test": []interface{}{"2"}},
			},
		)
	})
	t.Run(`with nodes node-2 self`, func(t *testing.T) {
		checkGet(t,
			ts2.URL+"/v1/route?test=2",
			200,
			map[string]interface{}{
				"id":     "node-2",
				"url":    "/v1/route",
				"params": map[string]interface{}{"test": []interface{}{"2"}},
			},
		)
	})
	t.Run(`with nodes node-1 proxy to node-2`, func(t *testing.T) {
		checkGet(t,
			ts1.URL+"/v1/route?node=node-2&test=2",
			200,
			map[string]interface{}{
				"id":     "node-2",
				"url":    "/v1/route",
				"params": map[string]interface{}{"test": []interface{}{"2"}},
			},
		)
	})
	t.Run(`with nodes node-1 redirect to node-2`, func(t *testing.T) {
		checkGet(t,
			ts1.URL+"/v1/route?node=node-2&redirect&test=2",
			200,
			map[string]interface{}{
				"id":     "node-2",
				"url":    "/v1/route",
				"params": map[string]interface{}{"test": []interface{}{"2"}},
			},
		)
	})
	t.Run(`configure node-2 to node-3`, func(t *testing.T) {
		nodesProducer.ConsumeMessage(bus.NewMessage("nodes", []interface{}{
			proto.NodeInfo{
				ID:        "node-1",
				Advertise: ts1.Listener.Addr().String(),
			},
			proto.NodeInfo{
				ID:        "node-3",
				Advertise: ts2.Listener.Addr().String(),
			},
		}))
	})
	t.Run(`with nodes node-1 proxy to node-2`, func(t *testing.T) {
		checkGet(t,
			ts1.URL+"/v1/route?node=node-2&test=2",
			404,
			nil,
		)
	})
	t.Run(`with nodes node-1 proxy to node-3`, func(t *testing.T) {
		checkGet(t,
			ts1.URL+"/v1/route?node=node-3&test=2",
			200,
			map[string]interface{}{
				"id":     "node-2",
				"url":    "/v1/route",
				"params": map[string]interface{}{"test": []interface{}{"2"}},
			},
		)
	})

}
