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
		assert.Equal(t, "### RESOURCE {\"Request\":{\"Name\":\"1\",\"Provider\":\"test\",\"Config\":{\"a\":1}},\"Values\":{\"a\":\"123\"}}\n", buf.String())
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
		assert.Equal(t, "### RESOURCE {\"Request\":{\"Name\":\"1\",\"Provider\":\"test\",\"Config\":{\"a\":1}}}\n", buf.String())
	})
}

func TestResource_UnmarshalLine(t *testing.T) {
	t.Run(`with values`, func(t *testing.T) {
		r := &allocation.Resource{}
		assert.NoError(t, r.UnmarshalItem("### RESOURCE {\"Request\":{\"Name\":\"1\",\"Provider\":\"test\",\"Config\":{\"a\":1}},\"Values\":{\"a\":\"123\"}}\n", allocation.SystemPaths{}))
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
	t.Run(`with values no config`, func(t *testing.T) {
		r := &allocation.Resource{}
		assert.NoError(t, r.UnmarshalItem("### RESOURCE {\"Request\":{\"Name\":\"1\",\"Provider\":\"test\"},\"Values\":{\"a\":\"123\"}}\n", allocation.SystemPaths{}))
		assert.Equal(t, &allocation.Resource{
			Request: manifest.Resource{
				Provider: "test",
				Name:     "1",
			},
			Values: map[string]string{
				"a": "123",
			},
		}, r)
	})
	t.Run(`without values`, func(t *testing.T) {
		r := &allocation.Resource{}
		assert.NoError(t, r.UnmarshalItem("### RESOURCE {\"Request\":{\"Name\":\"1\",\"Provider\":\"test\",\"Config\":{\"a\":1}}}\n", allocation.SystemPaths{}))
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

func TestResource_FromManifest(t *testing.T) {
	m := manifest.Resource{
		Name:     "1",
		Provider: "prov",
		Config: map[string]interface{}{
			"a": 1.0,
		},
	}
	t.Run(`with values`, func(t *testing.T) {
		r := &allocation.Resource{}
		err := r.FromManifest("test", m, manifest.FlatMap{
			"resource.test.1.__values": `
				{
					"allocated": "true",
					"b": "2"
				}
			`,
		})
		assert.NoError(t, err)
		assert.Equal(t, &allocation.Resource{
			Request: manifest.Resource{
				Provider: "prov",
				Name:     "1",
				Config: map[string]interface{}{
					"a": 1.0,
				},
			},
			Values: map[string]string{
				"allocated": "true",
				"b":         "2",
			},
		}, r)
	})
	t.Run(`without values`, func(t *testing.T) {
		r := &allocation.Resource{}
		err := r.FromManifest("test", m, manifest.FlatMap{})
		assert.NoError(t, err)
		assert.Equal(t, &allocation.Resource{
			Request: manifest.Resource{
				Provider: "prov",
				Name:     "1",
				Config: map[string]interface{}{
					"a": 1.0,
				},
			},
		}, r)
	})
}

func TestResourceSlice_Append(t *testing.T) {
	expect := allocation.ResourceSlice{
		&allocation.Resource{
			Request: manifest.Resource{
				Provider: "prov",
				Name:     "1",
				Config: map[string]interface{}{
					"a": 1.0,
				},
			},
		},
		&allocation.Resource{
			Request: manifest.Resource{
				Provider: "test",
				Name:     "2",
			},
		},
	}
	src := "### RESOURCE {\"Request\":{\"Name\":\"1\",\"Provider\":\"prov\",\"Config\":{\"a\":1}}}\n### RESOURCE {\"Request\":{\"Name\":\"2\",\"Provider\":\"test\"}}\n"
	t.Run(`restore`, func(t *testing.T) {
		var v allocation.ResourceSlice
		err := allocation.UnmarshalItemSlice(allocation.SystemPaths{}, &v, &allocation.Resource{}, src, []string{"### RESOURCE "})
		assert.NoError(t, err)
		assert.Equal(t, expect, v)
	})
}
