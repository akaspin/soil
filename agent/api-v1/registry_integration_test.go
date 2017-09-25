// +build ide test_integration

package api_v1_test

import (
	"bytes"
	"encoding/json"
	"github.com/akaspin/soil/api/api-v1-types"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
)

func Test_Integration_RegistryPut_Process(t *testing.T) {
	t.SkipNow()
	r, err := os.Open("testdata/example-multi.hcl")
	assert.NoError(t, err)
	defer r.Close()

	var pods manifest.Registry
	err = (&pods).Unmarshal("public", r)
	assert.NoError(t, err)

	data, err := json.Marshal(&pods)
	assert.NoError(t, err)

	req, err := http.NewRequest("PUT", "http://127.0.0.1:7651/v1/registry", bytes.NewReader(data))
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
