// +build ide test_unit

package api_v1_test

import (
	"context"
	"encoding/json"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/api-v1"
	"github.com/akaspin/soil/api"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestDrainEndpoint_Process(t *testing.T) {
	log := logx.GetLog("test")
	ctx := context.Background()

	var active int32
	var inactive int32
	drainFn := func(state bool) {
		if state {
			atomic.AddInt32(&active, 1)
		} else {
			atomic.AddInt32(&inactive, 1)
		}
	}

	router := api.NewRouter()
	router.Put("/v1/agent/drain", api_v1.NewDrainPutEndpoint(drainFn))
	router.Delete("/v1/agent/drain", api_v1.NewDrainDeleteEndpoint(drainFn))
	router.Get("/v1/agent/drain", api_v1.NewDrainGetEndpoint("test", func() bool {
		return true
	}))

	mux := http.NewServeMux()
	router.Bind(ctx, log, mux)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	t.Run("put", func(t *testing.T) {
		req, err := http.NewRequest("PUT", ts.URL+"/v1/agent/drain", nil)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, 200)
		assert.Equal(t, active, int32(1))
	})
	t.Run("delete", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", ts.URL+"/v1/agent/drain", nil)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, 200)
		assert.Equal(t, active, int32(1))
		assert.Equal(t, inactive, int32(1))
	})
	t.Run("get", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/v1/agent/drain")
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, 200)

		var res api_v1.DrainResponse
		err = json.NewDecoder(resp.Body).Decode(&res)
		assert.NoError(t, err)
		assert.Equal(t, res, api_v1.DrainResponse{
			AgentId: "test",
			Drain:   true,
		})
	})
}
