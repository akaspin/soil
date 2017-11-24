// build ide test_unit

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
	"reflect"
	"strings"
	"testing"
)

func TestRegistryPodsPutProcessor_Process(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cons := bus.NewTestingConsumer(ctx)
	endpoint := api.NewRegistryPodsPut(logx.GetLog("test"), cons)
	router := api_server.NewRouter(logx.GetLog("test"), endpoint)
	srv := httptest.NewServer(router)
	defer srv.Close()

	t.Run(`empty`, func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/v1/registry", srv.URL), strings.NewReader("[]"))
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, resp.StatusCode, 400)
	})
	t.Run(`upload`, func(t *testing.T) {
		v := manifest.Registry{
			{
				Name:      "1",
				Namespace: manifest.PublicNamespace,
			},
			{
				Name:      "2",
				Namespace: manifest.PublicNamespace,
			},
		}
		buf := &bytes.Buffer{}
		assert.NoError(t, json.NewEncoder(buf).Encode(v))

		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/v1/registry", srv.URL), bytes.NewReader(buf.Bytes()))
		require.NoError(t, err)
		_, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)

		fixture.WaitNoError10(t, cons.ExpectMessagesFn(
			bus.NewMessage("1", manifest.Pod{
				Name:      "1",
				Namespace: manifest.PublicNamespace,
			}),
			bus.NewMessage("2", manifest.Pod{
				Name:      "2",
				Namespace: manifest.PublicNamespace,
			}),
		))
	})
}

func TestRegistryPodsDeleteProcessor_Process(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cons := bus.NewTestingConsumer(ctx)
	endpoint := api.NewRegistryPodsDelete(logx.GetLog("test"), cons)
	router := api_server.NewRouter(logx.GetLog("test"), endpoint)
	srv := httptest.NewServer(router)
	defer srv.Close()

	t.Run(`empty query`, func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/v1/registry", srv.URL), nil)
		assert.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		require.NotNil(t, req)
		assert.Equal(t, resp.StatusCode, 400)
	})
	t.Run(`with two pods`, func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/v1/registry?pods=1&pods=2", srv.URL), nil)
		require.NoError(t, err)
		_, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)

		fixture.WaitNoError10(t, cons.ExpectMessagesFn(
			bus.NewMessage("1", nil),
			bus.NewMessage("2", nil),
		))
	})
}

func TestRegistryPodsGetProcessor_Process(t *testing.T) {
	endpoint := api.NewRegistryPodsGet()
	router := api_server.NewRouter(logx.GetLog("test"), endpoint)
	srv := httptest.NewServer(router)
	defer srv.Close()

	t.Run(`empty`, func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s/v1/registry", srv.URL))
		require.NoError(t, err)
		var pods manifest.Registry
		assert.NoError(t, json.NewDecoder(resp.Body).Decode(&pods))
		defer resp.Body.Close()
		assert.Equal(t, manifest.Registry{}, pods)
	})
	t.Run(`not empty`, func(t *testing.T) {
		v := manifest.Registry{
			{
				Name:      "1",
				Namespace: manifest.PublicNamespace,
			},
			{
				Name:      "2",
				Namespace: manifest.PublicNamespace,
			},
		}
		endpoint.Processor().(bus.Consumer).ConsumeMessage(bus.NewMessage("public", v))
		fixture.WaitNoError10(t, func() (err error) {
			var resp *http.Response
			if resp, err = http.Get(fmt.Sprintf("%s/v1/registry", srv.URL)); err != nil {
				return
			}
			if resp == nil {
				err = fmt.Errorf(`no response`)
				return
			}
			var pods manifest.Registry
			if err = json.NewDecoder(resp.Body).Decode(&pods); err != nil {
				return
			}
			defer resp.Body.Close()
			if !reflect.DeepEqual(v, pods) {
				err = fmt.Errorf(`bad response: %v`, pods)
			}
			return
		})
	})
}
