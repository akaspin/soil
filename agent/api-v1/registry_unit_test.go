// +build ide test_unit

package api_v1_test

import (
	"testing"
	"github.com/akaspin/soil/manifest"
	"os"
	"github.com/stretchr/testify/assert"
	"encoding/json"
	"net/http"
	"bytes"
	"github.com/akaspin/soil/api"
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/api-v1"
	"github.com/akaspin/soil/api/api-v1-types"
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

	var pods manifest.Pods
	err = (&pods).Unmarshal("public", r)
	assert.NoError(t, err)

	data, err := json.Marshal(&pods)
	assert.NoError(t, err)

	req, err := http.NewRequest("PUT", "http://127.0.0.1:8080/v1/pods", bytes.NewReader(data))
	assert.NoError(t, err)

	resp, err := (&http.Client{}).Do(req)
	assert.NoError(t, err)

	var marks api_v1_types.RegistrySubmitResponse
	err = json.NewDecoder(resp.Body).Decode(&marks)

	assert.Equal(t, marks, api_v1_types.RegistrySubmitResponse{
		"first": 0x515d988d1de74877,
		"second": 0x20a35ccef17a69c5,
	})
}
