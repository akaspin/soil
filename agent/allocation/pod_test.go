// +build ide test_unit

package allocation_test

import (
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/lib"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewFromManifest(t *testing.T) {

	env := map[string]string{
		"meta.consul":     "true",
		"system.pod_exec": "ExecStart=/usr/bin/sleep inf",
	}

	t.Run("0 simple names", func(t *testing.T) {
		var buffers lib.StaticBuffers
		assert.NoError(t, buffers.ReadFiles("testdata/test_new_from_manifest_0.hcl"))
		var pods manifest.Pods
		assert.NoError(t, pods.Unmarshal("private", buffers.GetReaders()...))

		m := pods[0]
		res := allocation.NewPod(allocation.DefaultSystemPaths())
		assert.NoError(t, res.FromManifest(m, env))

		assert.Equal(t, &allocation.Pod{
			Header: allocation.Header{
				Name:      "pod-1",
				PodMark:   5028014810488605184,
				AgentMark: 0x623669d2cde83725,
				Namespace: "private"},
			UnitFile: allocation.UnitFile{
				SystemPaths: allocation.DefaultSystemPaths(),
				Path:        "/run/systemd/system/pod-private-pod-1.service",
				Source:      "### POD pod-1 {\"AgentMark\":7076960218577909541,\"Namespace\":\"private\",\"PodMark\":5028014810488605184}\n### UNIT /run/systemd/system/unit-1.service {\"Create\":\"start\",\"Update\":\"\",\"Destroy\":\"stop\",\"Permanent\":false}\n### UNIT /run/systemd/system/unit-2.service {\"Create\":\"start\",\"Update\":\"\",\"Destroy\":\"stop\",\"Permanent\":false}\n### BLOB /etc/test {\"Leave\":false,\"Permissions\":420}\n\n[Unit]\nDescription=pod-1\nBefore=unit-1.service unit-2.service\n[Service]\nExecStart=/usr/bin/sleep inf\n[Install]\nWantedBy=multi-user.target\n"},
			Units: []*allocation.Unit{
				{
					UnitFile: allocation.UnitFile{
						SystemPaths: allocation.DefaultSystemPaths(),
						Path:        "/run/systemd/system/unit-1.service",
						Source:      "# true"},
					Transition: manifest.Transition{Create: "start", Update: "", Destroy: "stop", Permanent: false},
				},
				{
					UnitFile: allocation.UnitFile{
						SystemPaths: allocation.DefaultSystemPaths(),
						Path:        "/run/systemd/system/unit-2.service",
						Source:      "# true 10090666253179731817"},
					Transition: manifest.Transition{Create: "start", Update: "", Destroy: "stop", Permanent: false},
				},
			},
			Blobs: []*allocation.Blob{
				{Name: "/etc/test", Permissions: 420, Leave: false, Source: "test"},
			},
		}, res)
	})
	t.Run("interpolate names", func(t *testing.T) {
		var buffers lib.StaticBuffers
		var pods manifest.Pods
		assert.NoError(t, buffers.ReadFiles("testdata/test_new_from_manifest_1.hcl"))
		assert.NoError(t, pods.Unmarshal("private", buffers.GetReaders()...))

		m := pods[0]
		res := allocation.NewPod(allocation.DefaultSystemPaths())
		assert.NoError(t, res.FromManifest(m, env))

		assert.Equal(t, &allocation.Pod{
			Header: allocation.Header{
				Name:      "pod-2",
				PodMark:   17122768065259164291,
				AgentMark: 0x623669d2cde83725,
				Namespace: "private",
			},
			UnitFile: allocation.UnitFile{
				SystemPaths: allocation.DefaultSystemPaths(),
				Path:        "/run/systemd/system/pod-private-pod-2.service",
				Source:      "### POD pod-2 {\"AgentMark\":7076960218577909541,\"Namespace\":\"private\",\"PodMark\":17122768065259164291}\n### UNIT /run/systemd/system/private-unit-2.service {\"Create\":\"start\",\"Update\":\"\",\"Destroy\":\"stop\",\"Permanent\":false}\n### UNIT /run/systemd/system/pod-2-unit-1.service {\"Create\":\"start\",\"Update\":\"\",\"Destroy\":\"stop\",\"Permanent\":false}\n### BLOB /pod-2/etc/test {\"Leave\":false,\"Permissions\":420}\n\n[Unit]\nDescription=pod-2\nBefore=private-unit-2.service pod-2-unit-1.service\n[Service]\nExecStart=/usr/bin/sleep inf\n[Install]\nWantedBy=multi-user.target\n"},
			Units: []*allocation.Unit{
				{
					UnitFile: allocation.UnitFile{
						SystemPaths: allocation.DefaultSystemPaths(),
						Path:        "/run/systemd/system/private-unit-2.service",
						Source:      "# true 10090666253179731817"},
					Transition: manifest.Transition{Create: "start", Update: "", Destroy: "stop", Permanent: false},
				},
				{
					UnitFile: allocation.UnitFile{
						SystemPaths: allocation.DefaultSystemPaths(),
						Path:        "/run/systemd/system/pod-2-unit-1.service",
						Source:      "# true multi-user.target"},
					Transition: manifest.Transition{Create: "start", Update: "", Destroy: "stop", Permanent: false},
				},
			},
			Blobs: []*allocation.Blob{
				{Name: "/pod-2/etc/test", Permissions: 420, Leave: false, Source: "test"},
			},
		}, res)
	})
	t.Run("with resources", func(t *testing.T) {
		env3 := map[string]string{
			"meta.consul":                          "true",
			"system.pod_exec":                      "ExecStart=/usr/bin/sleep inf",
			"resource.port.pod-1.8080.__values":    `{"value":"8080"}`,
			"resource.counter.pod-1.main.__values": `{"value":"1"}`,
		}

		var buffers lib.StaticBuffers
		var pods manifest.Pods
		assert.NoError(t, buffers.ReadFiles("testdata/test_new_from_manifest_2.hcl"))
		assert.NoError(t, pods.Unmarshal("private", buffers.GetReaders()...))

		m := pods[0]
		res := allocation.NewPod(allocation.DefaultSystemPaths())
		err := res.FromManifest(m, env3)
		assert.NoError(t, err)
		assert.Equal(t, &allocation.Pod{
			Header: allocation.Header{
				Name:      "pod-1",
				PodMark:   5133150592239279453,
				AgentMark: 17576127034913539037,
				Namespace: "private",
			},
			UnitFile: allocation.UnitFile{
				SystemPaths: allocation.SystemPaths{
					Local:   "/etc/systemd/system",
					Runtime: "/run/systemd/system",
				},
				Path:   "/run/systemd/system/pod-private-pod-1.service",
				Source: "### POD pod-1 {\"AgentMark\":17576127034913539037,\"Namespace\":\"private\",\"PodMark\":5133150592239279453}\n### RESOURCE port 8080 {\"Request\":{\"fixed\":8080},\"Values\":{\"value\":\"8080\"}}\n### RESOURCE counter main {\"Request\":{\"count\":3},\"Values\":{\"value\":\"1\"}}\n\n[Unit]\nDescription=pod-1\nBefore=\n[Service]\nExecStart=/usr/bin/sleep inf\n[Install]\nWantedBy=multi-user.target\n",
			},
			Units: nil,
			Blobs: nil,
			Resources: []*allocation.Resource{
				{
					Request: manifest.Resource{
						Name: "8080",
						Kind: "port",
						Config: map[string]interface{}{
							"fixed": int(8080),
						},
					},
					Values: map[string]string{"value": "8080"},
				},
				{
					Request: manifest.Resource{
						Name: "main",
						Kind: "counter",
						Config: map[string]interface{}{
							"count": int(3),
						},
					},
					Values: map[string]string{"value": "1"},
				},
			},
		}, res)
	})
}

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
### PROVIDER {"Nature":"test","Kind":"test","Config":{"a":1,"b":"aa \"bb\""}}
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
		Providers: allocation.Providers{
			{
				Nature: "test",
				Kind:   "test",
				Config: map[string]interface{}{
					"a": float64(1),
					"b": `aa "bb"`,
				},
			},
		},
	},
		alloc)
}
