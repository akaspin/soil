// +build ide test_unit

package cluster_test

import (
	"github.com/akaspin/soil/agent/cluster"
	"github.com/akaspin/soil/lib"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConfig_Unmarshal(t *testing.T) {
	t.Run("one file", func(t *testing.T) {
		var buffers lib.StaticBuffers
		assert.NoError(t, buffers.ReadFiles("testdata/config_test_0.hcl"))
		var config cluster.Config
		assert.NoError(t, (&config).Unmarshal(buffers.GetReaders()...))
		assert.Equal(t, cluster.Config{
			NodeID:        "node-1-add",
			BackendURL:    "consul://127.0.0.1:8500",
			Advertise:     "127.0.0.1:7654",
			TTL:           time.Minute * 11,
			RetryInterval: time.Second * 30,
		}, config)
	})
}
