// +build ide test_unit

package api_test

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/api"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/fixture"
	"github.com/akaspin/soil/proto"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestClusterNodesProcessor_Process(t *testing.T) {
	processor := api.NewClusterNodesGet(logx.GetLog("test")).Processor()

	nodes := proto.NodesInfo{
		{
			ID:        "1",
			Advertise: "one",
			Version:   "0.1",
			API:       "v1",
		},
		{
			ID:        "2",
			Advertise: "two",
			Version:   "0.1",
			API:       "v1",
		},
	}

	t.Run(`empty`, func(t *testing.T) {
		res, _ := processor.Process(context.Background(), nil, nil)
		assert.Nil(t, res)
	})
	t.Run(`with nodes`, func(t *testing.T) {
		processor.(bus.Consumer).ConsumeMessage(bus.NewMessage("1", nodes))
		fixture.WaitNoError10(t, func() error {
			res, _ := processor.Process(context.Background(), nil, nil)
			if !reflect.DeepEqual(res, nodes) {
				return fmt.Errorf(`not equal`)
			}
			return nil
		})
	})
}
