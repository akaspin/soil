// +build ide test_unit

package allocation_test

import (
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewFromFS(t *testing.T) {

	paths := allocation.SystemPaths{
		Local:   "testdata/etc",
		Runtime: "testdata",
	}
	alloc := allocation.NewPod(paths)
	err := alloc.FromFilesystem("testdata/pod-test-1.service")
	assert.NoError(t, err)
	assert.Equal(t, &allocation.Pod{
		Header: allocation.Header{
			Name:      "test-1",
			PodMark:   123,
			AgentMark: 456,
			Namespace: "private",
		},
		UnitFile: allocation.UnitFile{
			SystemPaths: paths,
			Path:        "testdata/pod-test-1.service",
			Source: `### POD test-1 {"AgentMark":456,"Namespace":"private","PodMark":123}
### UNIT testdata/test-1-0.service {"Create":"start","Destroy":"stop","Permanent":true,"Update":"restart"}
### UNIT testdata/test-1-1.service {"Create":"start","Destroy":"stop","Permanent":true,"Update":"restart"}
### PROVIDER {"Kind":"test","Name":"test","Config":{"a":1,"b":"aa \"bb\""}}
### RESOURCE {"Request":{"Name":"8080","Provider":"pod-1.port","Config":{"a":1}},"Values":{"value":"9000"}}
[Unit]
Description=test-1
Before=test-1-0.service test-1-1.service
[Service]
ExecStart=/usr/bin/sleep inf
[Install]
WantedBy=multi-user.target
`,
		},
		Units: []*allocation.Unit{
			{
				UnitFile: allocation.UnitFile{
					SystemPaths: paths,
					Path:        "testdata/test-1-0.service",
					Source: `[Unit]
Description=Unit test-1-0.service
[Service]
ExecStart=/usr/bin/sleep inf
[Install]
WantedBy=multi-user.target
`,
				},
				Transition: manifest.Transition{
					Create:    "start",
					Update:    "restart",
					Destroy:   "stop",
					Permanent: true,
				},
			},
			{
				UnitFile: allocation.UnitFile{
					SystemPaths: paths,
					Path:        "testdata/test-1-1.service",
					Source: `[Unit]
Description=Unit test-1-1.service
[Service]
ExecStart=/usr/bin/sleep inf
[Install]
WantedBy=multi-user.target
`,
				},
				Transition: manifest.Transition{
					Create:    "start",
					Update:    "restart",
					Destroy:   "stop",
					Permanent: true,
				},
			},
		},
		Providers: allocation.ProviderSlice{
			{
				Kind: "test",
				Name: "test",
				Config: map[string]interface{}{
					"a": float64(1),
					"b": `aa "bb"`,
				},
			},
		},
		Resources: allocation.ResourceSlice{
			{
				Request: manifest.Resource{
					Name:     "8080",
					Provider: "pod-1.port",
					Config: map[string]interface{}{
						"a": 1.0,
					},
				},
				Values: manifest.FlatMap{
					//"allocated": "true",
					"value": "9000",
				},
			},
		},
	},
		alloc)
}
