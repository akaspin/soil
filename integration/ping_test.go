package integration_test

import (
	"testing"
	"net/http"
	"github.com/stretchr/testify/assert"
	"github.com/akaspin/soil/fixture"
)

func TestPing(t *testing.T) {
	fixture.RunTestIf(t, "TEST_INTEGRATION")
	resp, err := http.Get("http://127.0.0.1:7651/v1/status/ping")
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
}
