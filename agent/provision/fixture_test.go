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
		alloc, _ := allocation.NewFromManifest(pod, allocation.DefaultSystemDPaths(), map[string]string{
			"system.pod_exec": "ExecStart=/usr/bin/sleep inf",
		})
		recovered = append(recovered, alloc)
	}
	return
}
