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
	req, err := http.NewRequest(http.MethodPut, "http://127.0.0.1:7654/v1/agent/reload", nil)
	assert.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	// ping
	time.Sleep(time.Second)
	resp, err = http.Get("http://127.0.0.1:7654/v1/status/ping")
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	// drain
	time.Sleep(time.Second)
	req, err = http.NewRequest(http.MethodPut, "http://127.0.0.1:7654/v1/agent/drain", nil)
	assert.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	time.Sleep(time.Second)
	req, err = http.NewRequest(http.MethodDelete, "http://127.0.0.1:7654/v1/agent/drain", nil)
	assert.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	time.Sleep(time.Second)
	stopRequest, err := http.NewRequest(http.MethodPut, "http://127.0.0.1:7654/v1/agent/stop", nil)
	resp, err = http.DefaultClient.Do(stopRequest)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	wd.Wait()
}
