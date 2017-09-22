// +build ide test_integration

package api_v1

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestIntegration_StatusNode_Process(t *testing.T) {
	// first node
	resp, err := http.Get("http://127.0.0.1:7651/v1/status/node")
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
	var res map[string]map[string]string
	err = json.NewDecoder(resp.Body).Decode(&res)
	assert.Equal(t, res["agent"]["id"], "node-1.node.dc1.consul")
}
