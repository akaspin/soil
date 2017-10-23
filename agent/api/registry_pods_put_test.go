// +build ide test_unit

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/api"
	"github.com/akaspin/soil/agent/api/api-server"
	"github.com/akaspin/soil/lib"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/soil/proto"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegistryPodsPutProcessor_Process(t *testing.T) {
	log := logx.GetLog("test")
	backend := newFixtureBackend()
	processor := api.NewRegistryPodsPut(log, backend).Processor()

	var buffers lib.StaticBuffers
	var pods manifest.Registry
	assert.NoError(t, buffers.ReadFiles("testdata/example-multi.hcl"))
	assert.NoError(t, pods.Unmarshal("public", buffers.GetReaders()...))

	req := proto.RegistryPodsPutRequest(pods)
	_, err := processor.Process(context.Background(), nil, &req)
	assert.NoError(t, err)
	assert.Equal(t, backend.states, []map[string]string{
		map[string]string{
			"api-test-0": "{\"Namespace\":\"public\",\"Name\":\"api-test-0\",\"Runtime\":true,\"Target\":\"multi-user.target\",\"Constraint\":null,\"Units\":null,\"Blobs\":null,\"Resources\":null}",
			"api-test-1": "{\"Namespace\":\"public\",\"Name\":\"api-test-1\",\"Runtime\":true,\"Target\":\"multi-user.target\",\"Constraint\":{\"never\":\"deploy\"},\"Units\":null,\"Blobs\":null,\"Resources\":null}"}})
}

func TestNewRegistryPodsPut(t *testing.T) {
	log := logx.GetLog("test")
	backend := newFixtureBackend()
	endpoint := api.NewRegistryPodsPut(log, backend)

	ts := httptest.NewServer(api_server.NewRouter(log, endpoint))
	defer ts.Close()

	var buffers lib.StaticBuffers
	var pods manifest.Registry
	assert.NoError(t, buffers.ReadFiles("testdata/example-multi.hcl"))
	assert.NoError(t, pods.Unmarshal("public", buffers.GetReaders()...))

	data, err := json.Marshal(&pods)
	assert.NoError(t, err)

	request, err := http.NewRequest("PUT", fmt.Sprintf("%s%s", ts.URL, proto.V1RegistryPods), bytes.NewReader(data))
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(request)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	assert.Equal(t, backend.states, []map[string]string{
		map[string]string{
			"api-test-0": "{\"Namespace\":\"public\",\"Name\":\"api-test-0\",\"Runtime\":true,\"Target\":\"multi-user.target\",\"Constraint\":null,\"Units\":null,\"Blobs\":null,\"Resources\":null}",
			"api-test-1": "{\"Namespace\":\"public\",\"Name\":\"api-test-1\",\"Runtime\":true,\"Target\":\"multi-user.target\",\"Constraint\":{\"never\":\"deploy\"},\"Units\":null,\"Blobs\":null,\"Resources\":null}"}})
}
