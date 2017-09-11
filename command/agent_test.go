// +build ide test_systemd

package command_test

import (
	"github.com/akaspin/soil/command"
	"github.com/akaspin/soil/fixture"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
	"github.com/akaspin/soil/agent/api-v1"
	"encoding/json"
)

func TestAgent_Run_Stop(t *testing.T) {

	sd := fixture.NewSystemd("/run/systemd/system", "pod")
	defer sd.Cleanup()

	wd := &sync.WaitGroup{}
	wd.Add(1)

	go func() {
		defer wd.Done()
		err := command.Run(os.Stderr, os.Stdout, os.Stdin, []string{
			"agent", "--id", "node", "--meta", "rack=left", "--meta", "dc=1",
		}...)
		require.NoError(t, err)
	}()

	time.Sleep(time.Second)

	// reload
	resp, err := http.Get("http://127.0.0.1:7654/v1/agent/reload")
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	// ping
	time.Sleep(time.Second)
	resp, err = http.Get("http://127.0.0.1:7654/v1/status/ping")
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	// info
	resp, err = http.Get("http://127.0.0.1:7654/v1/status/info")
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
	var res1 api_v1.StatusInfoResponse
	err = json.NewDecoder(resp.Body).Decode(&res1)
	assert.NoError(t, err)
	assert.Equal(t, res1, api_v1.StatusInfoResponse{
		"agent": {
			Active: true,
			Namespaces: []string{"private", "public"},
			Data: map[string]string{
				"id": "node",
				"pod_exec": "ExecStart=/usr/bin/sleep inf",
			},
		},
		"allocation": {
			Active: true,
			Namespaces: []string{"private", "public"},
			Data: map[string]string{
			},
		},
		"meta": {
			Active: true,
			Namespaces: []string{"private", "public"},
			Data: map[string]string{
				"rack": "left",
				"dc": "1",
			},
		},
	})

	resp, err = http.Get("http://127.0.0.1:7654/v1/agent/stop")
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	wd.Wait()
}
