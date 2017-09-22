// +build ide test_integration

package api_v1_test

import (
	"encoding/json"
	"github.com/akaspin/soil/api/api-v1-types"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestPing(t *testing.T) {
	resp, err := http.Get("http://127.0.0.1:7651/v1/status/ping")
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestClusterNodesGET_Process(t *testing.T) {
	// first node
	resp, err := http.Get("http://127.0.0.1:7651/v1/public/nodes")
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
	var res api_v1_types.NodesResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	assert.Len(t, res, 3)

	resp, err = http.Get("http://127.0.0.1:7651/v1/public/nodes?node=node-3.node.dc1.consul")
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
	var res2 api_v1_types.NodesResponse
	err = json.NewDecoder(resp.Body).Decode(&res2)
	assert.Len(t, res2, 3)
}