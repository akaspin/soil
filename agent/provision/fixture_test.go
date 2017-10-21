package provision_test

import (
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/manifest"
	"testing"
)

func makeAllocations(t *testing.T, path string) (recovered []*allocation.Pod) {
	t.Helper()
	var pods manifest.Registry
	err := pods.UnmarshalFiles("private", path)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	for _, pod := range pods {
		alloc := allocation.NewPod(allocation.DefaultSystemPaths())
		err = alloc.FromManifest(pod, map[string]string{
			"system.pod_exec": "ExecStart=/usr/bin/sleep inf",
		})
		if err != nil {
			t.Error(err)
			t.Fail()
		}
		recovered = append(recovered, alloc)
	}
	return
}
