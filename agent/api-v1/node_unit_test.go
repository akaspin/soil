// +build ide test_unit

package api_v1_test

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/api-v1"
	"github.com/akaspin/soil/agent/metadata"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestUnit_StatusNode_Process(t *testing.T) {

	e := api_v1.NewStatusNode(logx.GetLog("test"))
	go e.Sync(metadata.Message{
		Prefix: "a",
		Clean:  true,
		Data: map[string]string{
			"a-k1": "a-v1",
		},
	})
	go e.Sync(metadata.Message{
		Prefix: "b",
		Clean:  true,
		Data: map[string]string{
			"b-k1": "b-v1",
		},
	})
	time.Sleep(time.Millisecond * 200)

	res, err := e.Process(context.Background(), nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, res, map[string]map[string]string{
		"a": {"a-k1": "a-v1"},
		"b": {"b-k1": "b-v1"},
	})
}
