// +build ide test_unit

package allocation_test

import (
	"bytes"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResource_MarshalLine(t *testing.T) {
	t.Run(`with values`, func(t *testing.T) {
		res := &allocation.Resource{
			Request: manifest.Resource{
				Provider: "test",
				Name:     "1",
				Config: map[string]interface{}{
					"a": 1.0,
				},
			},
			Values: map[string]string{
				"a": "123",
			},
		}
		var buf bytes.Buffer
		assert.NoError(t, res.MarshalLine(&buf))
		assert.Equal(t, "### RESOURCE.V2 {\"Request\":{\"Name\":\"1\",\"Provider\":\"test\",\"Config\":{\"a\":1}},\"Values\":{\"a\":\"123\"}}\n", buf.String())
	})
	t.Run(`without values`, func(t *testing.T) {
		res := &allocation.Resource{
			Request: manifest.Resource{
				Provider: "test",
				Name:     "1",
				Config: map[string]interface{}{
					"a": 1.0,
				},
			},
		}
		var buf bytes.Buffer
		assert.NoError(t, res.MarshalLine(&buf))
		assert.Equal(t, "### RESOURCE.V2 {\"Request\":{\"Name\":\"1\",\"Provider\":\"test\",\"Config\":{\"a\":1}}}\n", buf.String())
	})
}

func TestResource_UnmarshalLine(t *testing.T) {
	t.Run(`with values`, func(t *testing.T) {
		r := &allocation.Resource{}
		assert.NoError(t, r.UnmarshalLine("### RESOURCE.V2 {\"Request\":{\"Name\":\"1\",\"Provider\":\"test\",\"Config\":{\"a\":1}},\"Values\":{\"a\":\"123\"}}\n"))
		assert.Equal(t, &allocation.Resource{
			Request: manifest.Resource{
				Provider: "test",
				Name:     "1",
				Config: map[string]interface{}{
					"a": 1.0,
				},
			},
			Values: map[string]string{
				"a": "123",
			},
		}, r)
	})
	t.Run(`without values`, func(t *testing.T) {
		r := &allocation.Resource{}
		assert.NoError(t, r.UnmarshalLine("### RESOURCE.V2 {\"Request\":{\"Name\":\"1\",\"Provider\":\"test\",\"Config\":{\"a\":1}}}\n"))
		assert.Equal(t, &allocation.Resource{
			Request: manifest.Resource{
				Provider: "test",
				Name:     "1",
				Config: map[string]interface{}{
					"a": 1.0,
				},
			},
		}, r)
	})
}
