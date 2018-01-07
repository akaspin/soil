package provision_test

import (
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/lib"
	"github.com/akaspin/soil/manifest"
	"testing"
)

func makeAllocations(t *testing.T, path string) (recovered []*allocation.Pod) {
	t.Helper()
	var buffers lib.StaticBuffers
	var pods manifest.PodSlice
	if err := buffers.ReadFiles(path); err != nil {
		t.Error(err)
		t.Fail()
	}
	if err := pods.Unmarshal("private", buffers.GetReaders()...); err != nil {
		t.Error(err)
		t.Fail()
	}
	for _, pod := range pods {
		alloc := &allocation.Pod{
			UnitFile: allocation.UnitFile{
				SystemPaths: allocation.DefaultSystemPaths(),
			},
		}

		if err := alloc.FromManifest(pod, map[string]string{
			"system.pod_exec": "ExecStart=/usr/bin/sleep inf",
		}); err != nil {
			t.Error(err)
			t.Fail()
		}
		recovered = append(recovered, alloc)
	}
	return
}
