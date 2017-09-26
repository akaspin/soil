package api_v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/api-v1"
	"github.com/akaspin/soil/agent/api-v1/api-server"
	"github.com/akaspin/soil/proto"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegistryPodsDeleteProcessor_Process(t *testing.T) {
	backend := newFixtureBackend()
	processor := api_v1.NewRegistryPodsDelete(backend).Processor()

	processor.Process(context.Background(), nil, &proto.RegistryPodsDeleteRequest{
		"1", "2",
	})
	assert.Equal(t, backend.states, []map[string]string{
		{
			"1": "DELETE",
			"2": "DELETE",
		}})
}

func TestNewRegistryPodsDelete(t *testing.T) {
	log := logx.GetLog("test")
	backend := newFixtureBackend()
	endpoint := api_v1.NewRegistryPodsDelete(backend)

	ts := httptest.NewServer(api_server.NewRouter(log, endpoint))
	defer ts.Close()

	ingest := proto.RegistryPodsDeleteRequest{"1", "2"}
	data, err := json.Marshal(&ingest)
	assert.NoError(t, err)

	request, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s%s", ts.URL, proto.V1RegistryPods), bytes.NewReader(data))
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(request)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
	assert.Equal(t, backend.states, []map[string]string{
		{
			"1": "DELETE",
			"2": "DELETE",
		}})
}
