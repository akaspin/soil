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
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/fixture"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegistryPodPutProcessor_Process(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cons := bus.NewTestingConsumer(ctx)
	endpoint := api.NewRegistryPodPut(cons)
	router := api_server.NewRouter(logx.GetLog("test"), endpoint)
	srv := httptest.NewServer(router)
	defer srv.Close()

	t.Run(`upload`, func(t *testing.T) {
		v := manifest.DefaultPod(manifest.PublicNamespace)
		v.Name = "test"
		buf := &bytes.Buffer{}
		assert.NoError(t, json.NewEncoder(buf).Encode(v))

		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/v1/registry/pod", srv.URL), bytes.NewReader(buf.Bytes()))
		assert.NoError(t, err)
		_, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)
		require.NotNil(t, req)

		fixture.WaitNoError10(t, cons.ExpectMessagesFn(
			bus.NewMessage("test", v),
		))
	})
}
