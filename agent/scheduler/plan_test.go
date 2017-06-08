package scheduler_test

import (
	"fmt"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPlanUnit(t *testing.T) {
	left := &scheduler.AllocationUnit{
		AllocationUnitHeader: &scheduler.AllocationUnitHeader{
			Transition: manifest.Transition{
				Create:  "start",
				Update:  "restart",
				Destroy: "stop",
			},
		},
		AllocationFile: &scheduler.AllocationFile{
			Path:   "/run/systemd/system/unit-1.service",
			Source: "unit-1-0",
		},
	}
	t.Run("destroy", func(t *testing.T) {
		res := scheduler.PlanUnit(left, nil)
		assert.Equal(t, "[1:remove:/run/systemd/system/unit-1.service 0:stop:/run/systemd/system/unit-1.service]", fmt.Sprint(res))
	})
	t.Run("create", func(t *testing.T) {
		res := scheduler.PlanUnit(nil, left)
		assert.Equal(t, "[2:write:/run/systemd/system/unit-1.service 3:disable:/run/systemd/system/unit-1.service 4:start:/run/systemd/system/unit-1.service]", fmt.Sprint(res))
	})
	t.Run("update", func(t *testing.T) {
		right := &scheduler.AllocationUnit{
			AllocationUnitHeader: &scheduler.AllocationUnitHeader{
				Transition: manifest.Transition{
					Create:  "start",
					Update:  "restart",
					Destroy: "stop",
				},
			},
			AllocationFile: &scheduler.AllocationFile{
				Path:   "/run/systemd/system/unit-1.service",
				Source: "unit-1-1",
			},
		}
		res := scheduler.PlanUnit(left, right)
		assert.Equal(t, "[2:write:/run/systemd/system/unit-1.service 3:disable:/run/systemd/system/unit-1.service 4:restart:/run/systemd/system/unit-1.service]", fmt.Sprint(res))
	})
	t.Run("runtime to local", func(t *testing.T) {
		right := &scheduler.AllocationUnit{
			AllocationUnitHeader: &scheduler.AllocationUnitHeader{
				Transition: manifest.Transition{
					Create:  "start",
					Update:  "restart",
					Destroy: "stop",
				},
			},
			AllocationFile: &scheduler.AllocationFile{
				Path:   "/etc/systemd/system/unit-1.service",
				Source: "unit-1-0",
			},
		}
		res := scheduler.PlanUnit(left, right)
		assert.Equal(t, "[1:remove:/run/systemd/system/unit-1.service 0:stop:/run/systemd/system/unit-1.service 2:write:/etc/systemd/system/unit-1.service 3:disable:/etc/systemd/system/unit-1.service 4:start:/etc/systemd/system/unit-1.service]", fmt.Sprint(res))
	})
}

func TestPlanBlob(t *testing.T) {
	left1 := &scheduler.AllocationBlob{
		Name:        "/etc/test",
		Permissions: 0644,
		Source:      "left1",
	}
	left2 := &scheduler.AllocationBlob{
		Name:        "/etc/test",
		Permissions: 0644,
		Leave:       true,
		Source:      "left2",
	}
	right1 := &scheduler.AllocationBlob{
		Name:        "/etc/test",
		Permissions: 0644,
		Source:      "right1",
	}
	right2 := &scheduler.AllocationBlob{
		Name:        "/etc/test",
		Permissions: 0755,
		Source:      "left1",
	}
	t.Run("destroy left 1", func(t *testing.T) {
		res := scheduler.PlanBlob(left1, nil)
		assert.Equal(t, "[2:blob-destroy:/etc/test]", fmt.Sprint(res))
	})
	t.Run("destroy left 2", func(t *testing.T) {
		res := scheduler.PlanBlob(left2, nil)
		assert.Equal(t, "[]", fmt.Sprint(res))
	})
	t.Run("write right 1", func(t *testing.T) {
		res := scheduler.PlanBlob(nil, right1)
		assert.Equal(t, "[2:blob-write:/etc/test]", fmt.Sprint(res))
	})
	t.Run("write right 1 over left 1", func(t *testing.T) {
		res := scheduler.PlanBlob(left1, right1)
		assert.Equal(t, "[2:blob-write:/etc/test]", fmt.Sprint(res))
	})
	t.Run("write right 2 over left 1", func(t *testing.T) {
		res := scheduler.PlanBlob(left1, right2)
		assert.Equal(t, "[2:blob-write:/etc/test]", fmt.Sprint(res))
	})
}

func TestPlan(t *testing.T) {
	left := &scheduler.Allocation{
		AllocationHeader: &scheduler.AllocationHeader{
			Name:      "pod-1",
			AgentMark: 123,
			PodMark:   456,
			Namespace: "private",
		},
		AllocationFile: &scheduler.AllocationFile{
			Path:   "/etc/systemd/system/pod-pod-1.service",
			Source: "fake",
		},
		Units: []*scheduler.AllocationUnit{
			{
				AllocationFile: &scheduler.AllocationFile{
					Path:   "/etc/systemd/system/unit-1.service",
					Source: "fake",
				},
				AllocationUnitHeader: &scheduler.AllocationUnitHeader{
					Permanent: true,
					Transition: manifest.Transition{
						Create:  "start",
						Update:  "restart",
						Destroy: "stop",
					},
				},
			},
			{
				AllocationFile: &scheduler.AllocationFile{
					Path:   "/etc/systemd/system/unit-2.service",
					Source: "fake",
				},
				AllocationUnitHeader: &scheduler.AllocationUnitHeader{
					Permanent: true,
					Transition: manifest.Transition{
						Create:  "start",
						Update:  "restart",
						Destroy: "stop",
					},
				},
			},
		},
		Blobs: []*scheduler.AllocationBlob{
			{
				Name:        "/etc/test1",
				Permissions: 0644,
				Source:      "test",
			},
		},
	}

	t.Run("noop pod", func(t *testing.T) {
		right := &scheduler.Allocation{
			AllocationHeader: &scheduler.AllocationHeader{
				Name:      "pod-1",
				AgentMark: 123,
				PodMark:   456,
				Namespace: "private",
			},
			AllocationFile: &scheduler.AllocationFile{
				Path:   "/etc/systemd/system/pod-pod-1.service",
				Source: "fake",
			},
			Units: []*scheduler.AllocationUnit{
				{
					AllocationFile: &scheduler.AllocationFile{
						Path:   "/etc/systemd/system/unit-1.service",
						Source: "fake",
					},
					AllocationUnitHeader: &scheduler.AllocationUnitHeader{
						Permanent: true,
						Transition: manifest.Transition{
							Create:  "start",
							Update:  "restart",
							Destroy: "stop",
						},
					},
				},
				{
					AllocationFile: &scheduler.AllocationFile{
						Path:   "/etc/systemd/system/unit-2.service",
						Source: "fake",
					},
					AllocationUnitHeader: &scheduler.AllocationUnitHeader{
						Permanent: true,
						Transition: manifest.Transition{
							Create:  "start",
							Update:  "restart",
							Destroy: "stop",
						},
					},
				},
			},
			Blobs: []*scheduler.AllocationBlob{
				{
					Name:        "/etc/test1",
					Permissions: 0644,
					Source:      "test",
				},
			},
		}
		assert.Equal(t, "[]", fmt.Sprint(scheduler.Plan(left, right)))
	})
	t.Run("unit-1 perm to disabled", func(t *testing.T) {
		right := &scheduler.Allocation{
			AllocationHeader: &scheduler.AllocationHeader{
				Name:      "pod-1",
				AgentMark: 123,
				PodMark:   456,
				Namespace: "private",
			},
			AllocationFile: &scheduler.AllocationFile{
				Path:   "/etc/systemd/system/pod-pod-1.service",
				Source: "fake",
			},
			Units: []*scheduler.AllocationUnit{
				{
					AllocationFile: &scheduler.AllocationFile{
						Path:   "/etc/systemd/system/unit-1.service",
						Source: "fake",
					},
					AllocationUnitHeader: &scheduler.AllocationUnitHeader{
						Permanent: false,
						Transition: manifest.Transition{
							Create:  "start",
							Update:  "restart",
							Destroy: "stop",
						},
					},
				},
				{
					AllocationFile: &scheduler.AllocationFile{
						Path:   "/etc/systemd/system/unit-2.service",
						Source: "fake",
					},
					AllocationUnitHeader: &scheduler.AllocationUnitHeader{
						Permanent: true,
						Transition: manifest.Transition{
							Create:  "start",
							Update:  "restart",
							Destroy: "stop",
						},
					},
				},
			},
		}
		assert.Equal(t, "[2:blob-destroy:/etc/test1 3:disable:/etc/systemd/system/unit-1.service]", fmt.Sprint(scheduler.Plan(left, right)))
	})
	t.Run("update unit-1 and file", func(t *testing.T) {
		right := &scheduler.Allocation{
			AllocationHeader: &scheduler.AllocationHeader{
				Name:      "pod-1",
				AgentMark: 123,
				PodMark:   456,
				Namespace: "private",
			},
			AllocationFile: &scheduler.AllocationFile{
				Path:   "/etc/systemd/system/pod-pod-1.service",
				Source: "fake",
			},
			Units: []*scheduler.AllocationUnit{
				{
					AllocationFile: &scheduler.AllocationFile{
						Path:   "/etc/systemd/system/unit-1.service",
						Source: "fake1",
					},
					AllocationUnitHeader: &scheduler.AllocationUnitHeader{
						Permanent: true,
						Transition: manifest.Transition{
							Create:  "start",
							Update:  "restart",
							Destroy: "stop",
						},
					},
				},
				{
					AllocationFile: &scheduler.AllocationFile{
						Path:   "/etc/systemd/system/unit-2.service",
						Source: "fake",
					},
					AllocationUnitHeader: &scheduler.AllocationUnitHeader{
						Permanent: true,
						Transition: manifest.Transition{
							Create:  "start",
							Update:  "restart",
							Destroy: "stop",
						},
					},
				},
			},
			Blobs: []*scheduler.AllocationBlob{
				{
					Name:        "/etc/test1",
					Permissions: 0644,
					Source:      "test1",
				},
			},
		}
		assert.Equal(t, "[2:write:/etc/systemd/system/unit-1.service 2:blob-write:/etc/test1 3:enable:/etc/systemd/system/unit-1.service 4:restart:/etc/systemd/system/unit-1.service]", fmt.Sprint(scheduler.Plan(left, right)))
	})
	t.Run("create pod", func(t *testing.T) {
		assert.Equal(t, "[2:write:/etc/systemd/system/pod-pod-1.service 2:write:/etc/systemd/system/unit-1.service 2:write:/etc/systemd/system/unit-2.service 2:blob-write:/etc/test1 3:enable:/etc/systemd/system/pod-pod-1.service 3:enable:/etc/systemd/system/unit-1.service 3:enable:/etc/systemd/system/unit-2.service 4:start:/etc/systemd/system/pod-pod-1.service 4:start:/etc/systemd/system/unit-1.service 4:start:/etc/systemd/system/unit-2.service]", fmt.Sprint(scheduler.Plan(nil, left)))
	})
	t.Run("destroy pod", func(t *testing.T) {
		assert.Equal(t, "[0:stop:/etc/systemd/system/pod-pod-1.service 0:stop:/etc/systemd/system/unit-1.service 0:stop:/etc/systemd/system/unit-2.service 1:remove:/etc/systemd/system/pod-pod-1.service 1:remove:/etc/systemd/system/unit-1.service 1:remove:/etc/systemd/system/unit-2.service 2:blob-destroy:/etc/test1]", fmt.Sprint(scheduler.Plan(left, nil)))
	})
	t.Run("change prefix", func(t *testing.T) {
		right := &scheduler.Allocation{
			AllocationHeader: &scheduler.AllocationHeader{
				Name:      "pod-1",
				AgentMark: 123,
				PodMark:   456,
				Namespace: "private",
			},
			AllocationFile: &scheduler.AllocationFile{
				Path:   "/etc/systemd/system/pod-local-pod-1.service",
				Source: "fake",
			},
			Units: []*scheduler.AllocationUnit{
				{
					AllocationFile: &scheduler.AllocationFile{
						Path:   "/etc/systemd/system/unit-1.service",
						Source: "fake",
					},
					AllocationUnitHeader: &scheduler.AllocationUnitHeader{
						Permanent: true,
						Transition: manifest.Transition{
							Create:  "start",
							Update:  "restart",
							Destroy: "stop",
						},
					},
				},
				{
					AllocationFile: &scheduler.AllocationFile{
						Path:   "/etc/systemd/system/unit-2.service",
						Source: "fake",
					},
					AllocationUnitHeader: &scheduler.AllocationUnitHeader{
						Permanent: true,
						Transition: manifest.Transition{
							Create:  "start",
							Update:  "restart",
							Destroy: "stop",
						},
					},
				},
			},
			Blobs: []*scheduler.AllocationBlob{
				{
					Name:        "/etc/test1",
					Permissions: 0644,
					Source:      "test",
				},
			},
		}
		assert.Equal(t, "[0:stop:/etc/systemd/system/pod-pod-1.service 1:remove:/etc/systemd/system/pod-pod-1.service 2:write:/etc/systemd/system/pod-local-pod-1.service 3:enable:/etc/systemd/system/pod-local-pod-1.service 4:start:/etc/systemd/system/pod-local-pod-1.service]", fmt.Sprint(scheduler.Plan(left, right)))
	})
	t.Run("local to runtime", func(t *testing.T) {
		right := &scheduler.Allocation{
			AllocationHeader: &scheduler.AllocationHeader{
				Name:      "pod-1",
				AgentMark: 123,
				PodMark:   456,
				Namespace: "private",
			},
			AllocationFile: &scheduler.AllocationFile{
				Path:   "/run/systemd/system/pod-pod-1.service",
				Source: "fake",
			},
			Units: []*scheduler.AllocationUnit{
				{
					AllocationFile: &scheduler.AllocationFile{
						Path:   "/run/systemd/system/unit-1.service",
						Source: "fake",
					},
					AllocationUnitHeader: &scheduler.AllocationUnitHeader{
						Permanent: true,
						Transition: manifest.Transition{
							Create:  "start",
							Update:  "restart",
							Destroy: "stop",
						},
					},
				},
				{
					AllocationFile: &scheduler.AllocationFile{
						Path:   "/run/systemd/system/unit-2.service",
						Source: "fake",
					},
					AllocationUnitHeader: &scheduler.AllocationUnitHeader{
						Permanent: true,
						Transition: manifest.Transition{
							Create:  "start",
							Update:  "restart",
							Destroy: "stop",
						},
					},
				},
			},
			Blobs: []*scheduler.AllocationBlob{
				{
					Name:        "/etc/test1",
					Permissions: 0644,
					Source:      "test",
				},
			},
		}
		assert.Equal(t, "[0:stop:/etc/systemd/system/pod-pod-1.service 0:stop:/etc/systemd/system/unit-1.service 0:stop:/etc/systemd/system/unit-2.service 1:remove:/etc/systemd/system/pod-pod-1.service 1:remove:/etc/systemd/system/unit-1.service 1:remove:/etc/systemd/system/unit-2.service 2:write:/run/systemd/system/pod-pod-1.service 2:write:/run/systemd/system/unit-1.service 2:write:/run/systemd/system/unit-2.service 3:enable:/run/systemd/system/pod-pod-1.service 3:enable:/run/systemd/system/unit-1.service 3:enable:/run/systemd/system/unit-2.service 4:start:/run/systemd/system/pod-pod-1.service 4:start:/run/systemd/system/unit-1.service 4:start:/run/systemd/system/unit-2.service]", fmt.Sprint(scheduler.Plan(left, right)))
	})
}
