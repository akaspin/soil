// +build ide test_unit

package api_v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/api-v1"
	"github.com/akaspin/soil/api"
	"github.com/akaspin/soil/api/api-v1-types"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
)

func Test_Unit_RegistryPut_Process(t *testing.T) {
	ctx := context.Background()
	log := logx.GetLog("test")

	backend := newFixtureBackend()
	handler := api_v1.NewRegistryPut(log, backend)
	router := api.NewRouter(ctx, log,
		api.PUT("/v1/pods", handler),
	)

	server := api.NewServer(ctx, log, ":8080", router)
	assert.NoError(t, server.Open())
	defer server.Close()

	r, err := os.Open("testdata/example-multi.hcl")
	assert.NoError(t, err)
	defer r.Close()

	var pods manifest.Registry
	err = (&pods).Unmarshal("public", r)
	assert.NoError(t, err)

	data, err := json.Marshal(&pods)
	assert.NoError(t, err)

	req, err := http.NewRequest("PUT", "http://127.0.0.1:8080/v1/pods", bytes.NewReader(data))
	assert.NoError(t, err)

	resp, err := (&http.Client{}).Do(req)
	assert.NoError(t, err)

	var marks api_v1_types.RegistryPutResponse
	err = json.NewDecoder(resp.Body).Decode(&marks)

	assert.Equal(t, marks, api_v1_types.RegistryPutResponse{
		"api-test-0":0x937fb0d3a39c7353,
		"api-test-1":0x77cf87a45dfb307d,
	})
}
