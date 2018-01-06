// +build ide test_unit

package allocation_test

import (
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHeader(t *testing.T) {
	t.Skip()
	src := `### POD pod-1 {"AgentMark":123,"Namespace":"private","PodMark":345}
### UNIT /etc/systemd/system/unit-1.service {"Create":"start","Update":"","Destroy":"","Permanent":true}
### UNIT /etc/systemd/system/unit-2.service {"Create":"","Update":"","Destroy":"","Permanent":false}
### BLOB /etc/test {"Leave":false,"Permissions":420}
### RESOURCE port 8080 {"Request":{"fixed":8080,"other":"aaa bbb"},"Values":{"value":"8080"}}
### RESOURCE counter 1 {"Request":{},"Values":{"value":"1"}}
`
	expectUnits := []*allocation.Unit{
		{
			Transition: manifest.Transition{
				Create:    "start",
				Permanent: true,
			},
			UnitFile: allocation.UnitFile{
				SystemPaths: allocation.DefaultSystemPaths(),
				Path:        "/etc/systemd/system/unit-1.service",
			},
		},
		{
			Transition: manifest.Transition{
				Permanent: false,
			},
			UnitFile: allocation.UnitFile{
				SystemPaths: allocation.DefaultSystemPaths(),
				Path:        "/etc/systemd/system/unit-2.service",
			},
		},
	}
	expectBlobs := []*allocation.Blob{
		{
			Name:        "/etc/test",
			Permissions: 0644,
			Source:      "",
		},
	}
	expectResources := []*allocation.Resource{
		{
			Request: manifest.Resource{
				Provider: "port",
				Name:     "8080",
				Config: map[string]interface{}{
					"fixed": float64(8080),
					"other": "aaa bbb",
				},
			},
			Values: map[string]string{
				"value": "8080",
			},
		},
		{
			Request: manifest.Resource{
				Provider: "counter",
				Name:     "1",
				Config:   map[string]interface{}{},
			},
			Values: map[string]string{
				"value": "1",
			},
		},
	}
	expectHeader := &allocation.Header{
		Namespace: "private",
		AgentMark: 123,
		PodMark:   345,
	}
	t.Run("marshal", func(t *testing.T) {
		res, err := expectHeader.Marshal("pod-1", expectUnits, expectBlobs, expectResources, nil)
		assert.NoError(t, err)
		assert.Equal(t, res, src)
	})
	t.Run("unmarshal", func(t *testing.T) {
		header := &allocation.Header{}
		units, err := header.Unmarshal(src, allocation.DefaultSystemPaths())
		assert.NoError(t, err)
		assert.Equal(t, units, expectUnits)
		//assert.Equal(t, blobs, expectBlobs)
	})
}
