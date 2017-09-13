// +build ide test_integration

package api_v1_test

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestPing(t *testing.T) {
	resp, err := http.Get("http://127.0.0.1:7651/v1/status/ping")
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
	t.Log("OK")
}
