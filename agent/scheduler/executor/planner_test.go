package executor_test

import (
	"fmt"
	"github.com/akaspin/soil/agent/scheduler/allocation"
	"github.com/akaspin/soil/agent/scheduler/executor"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"testing"

)

func TestPlanUnit(t *testing.T) {
	left := &allocation.AllocationUnit{
		AllocationUnitHeader: &allocation.AllocationUnitHeader{
			Transition: manifest.Transition{
				Create:  "start",
				Update:  "restart",
				Destroy: "stop",
			},
		},
		AllocationFile: &allocation.AllocationFile{
			Path:   "/run/systemd/system/unit-1.service",
			Source: "unit-1-0",
		},
	}
	t.Run("destroy", func(t *testing.T) {
		res := executor.PlanUnit(left, nil)
		assert.Equal(t, "[1:remove:/run/systemd/system/unit-1.service 0:stop:/run/systemd/system/unit-1.service]", fmt.Sprint(res))
	})
	t.Run("create", func(t *testing.T) {
		res := executor.PlanUnit(nil, left)
		assert.Equal(t, "[2:write:/run/systemd/system/unit-1.service 3:disable:/run/systemd/system/unit-1.service 4:start:/run/systemd/system/unit-1.service]", fmt.Sprint(res))
	})
	t.Run("update", func(t *testing.T) {
		right := &allocation.AllocationUnit{
			AllocationUnitHeader: &allocation.AllocationUnitHeader{
				Transition: manifest.Transition{
					Create:  "start",
					Update:  "restart",
					Destroy: "stop",
				},
			},
			AllocationFile: &allocation.AllocationFile{
				Path:   "/run/systemd/system/unit-1.service",
				Source: "unit-1-1",
			},
		}
		res := executor.PlanUnit(left, right)
		assert.Equal(t, "[2:write:/run/systemd/system/unit-1.service 3:disable:/run/systemd/system/unit-1.service 4:restart:/run/systemd/system/unit-1.service]", fmt.Sprint(res))
	})
	t.Run("runtime to local", func(t *testing.T) {
		right := &allocation.AllocationUnit{
			AllocationUnitHeader: &allocation.AllocationUnitHeader{
				Transition: manifest.Transition{
					Create:  "start",
					Update:  "restart",
					Destroy: "stop",
				},
			},
			AllocationFile: &allocation.AllocationFile{
				Path:   "/etc/systemd/system/unit-1.service",
				Source: "unit-1-0",
			},
		}
		res := executor.PlanUnit(left, right)
		assert.Equal(t, "[1:remove:/run/systemd/system/unit-1.service 0:stop:/run/systemd/system/unit-1.service 2:write:/etc/systemd/system/unit-1.service 3:disable:/etc/systemd/system/unit-1.service 4:start:/etc/systemd/system/unit-1.service]", fmt.Sprint(res))
	})
}

func TestPlan(t *testing.T) {
	left := &allocation.Allocation{
		AllocationHeader: &allocation.AllocationHeader{
			Name:      "pod-1",
			AgentMark: 123,
			PodMark:   456,
			Namespace: "private",
		},
		AllocationFile: &allocation.AllocationFile{
			Path:   "/etc/systemd/system/pod-pod-1.service",
			Source: "fake",
		},
		Units: []*allocation.AllocationUnit{
			{
				AllocationFile: &allocation.AllocationFile{
					Path:   "/etc/systemd/system/unit-1.service",
					Source: "fake",
				},
				AllocationUnitHeader: &allocation.AllocationUnitHeader{
					Permanent: true,
					Transition: manifest.Transition{
						Create:  "start",
						Update:  "restart",
						Destroy: "stop",
					},
				},
			},
			{
				AllocationFile: &allocation.AllocationFile{
					Path:   "/etc/systemd/system/unit-2.service",
					Source: "fake",
				},
				AllocationUnitHeader: &allocation.AllocationUnitHeader{
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

	t.Run("noop pod", func(t *testing.T) {
		right := &allocation.Allocation{
			AllocationHeader: &allocation.AllocationHeader{
				Name:      "pod-1",
				AgentMark: 123,
				PodMark:   456,
				Namespace: "private",
			},
			AllocationFile: &allocation.AllocationFile{
				Path:   "/etc/systemd/system/pod-pod-1.service",
				Source: "fake",
			},
			Units: []*allocation.AllocationUnit{
				{
					AllocationFile: &allocation.AllocationFile{
						Path:   "/etc/systemd/system/unit-1.service",
						Source: "fake",
					},
					AllocationUnitHeader: &allocation.AllocationUnitHeader{
						Permanent: true,
						Transition: manifest.Transition{
							Create:  "start",
							Update:  "restart",
							Destroy: "stop",
						},
					},
				},
				{
					AllocationFile: &allocation.AllocationFile{
						Path:   "/etc/systemd/system/unit-2.service",
						Source: "fake",
					},
					AllocationUnitHeader: &allocation.AllocationUnitHeader{
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
		assert.Equal(t, "[]", fmt.Sprint(executor.Plan(left, right)))
	})
	t.Run("unit-1 perm to disabled", func(t *testing.T) {
		right := &allocation.Allocation{
			AllocationHeader: &allocation.AllocationHeader{
				Name:      "pod-1",
				AgentMark: 123,
				PodMark:   456,
				Namespace: "private",
			},
			AllocationFile: &allocation.AllocationFile{
				Path:   "/etc/systemd/system/pod-pod-1.service",
				Source: "fake",
			},
			Units: []*allocation.AllocationUnit{
				{
					AllocationFile: &allocation.AllocationFile{
						Path:   "/etc/systemd/system/unit-1.service",
						Source: "fake",
					},
					AllocationUnitHeader: &allocation.AllocationUnitHeader{
						Permanent: false,
						Transition: manifest.Transition{
							Create:  "start",
							Update:  "restart",
							Destroy: "stop",
						},
					},
				},
				{
					AllocationFile: &allocation.AllocationFile{
						Path:   "/etc/systemd/system/unit-2.service",
						Source: "fake",
					},
					AllocationUnitHeader: &allocation.AllocationUnitHeader{
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
		assert.Equal(t, "[3:disable:/etc/systemd/system/unit-1.service]", fmt.Sprint(executor.Plan(left, right)))
	})
	t.Run("update unit-1", func(t *testing.T) {
		right := &allocation.Allocation{
			AllocationHeader: &allocation.AllocationHeader{
				Name:      "pod-1",
				AgentMark: 123,
				PodMark:   456,
				Namespace: "private",
			},
			AllocationFile: &allocation.AllocationFile{
				Path:   "/etc/systemd/system/pod-pod-1.service",
				Source: "fake",
			},
			Units: []*allocation.AllocationUnit{
				{
					AllocationFile: &allocation.AllocationFile{
						Path:   "/etc/systemd/system/unit-1.service",
						Source: "fake1",
					},
					AllocationUnitHeader: &allocation.AllocationUnitHeader{
						Permanent: true,
						Transition: manifest.Transition{
							Create:  "start",
							Update:  "restart",
							Destroy: "stop",
						},
					},
				},
				{
					AllocationFile: &allocation.AllocationFile{
						Path:   "/etc/systemd/system/unit-2.service",
						Source: "fake",
					},
					AllocationUnitHeader: &allocation.AllocationUnitHeader{
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
		assert.Equal(t, "[2:write:/etc/systemd/system/unit-1.service 3:enable:/etc/systemd/system/unit-1.service 4:restart:/etc/systemd/system/unit-1.service]", fmt.Sprint(executor.Plan(left, right)))
	})
	t.Run("create pod", func(t *testing.T) {
		assert.Equal(t, "[2:write:/etc/systemd/system/pod-pod-1.service 2:write:/etc/systemd/system/unit-1.service 2:write:/etc/systemd/system/unit-2.service 3:enable:/etc/systemd/system/pod-pod-1.service 3:enable:/etc/systemd/system/unit-1.service 3:enable:/etc/systemd/system/unit-2.service 4:start:/etc/systemd/system/pod-pod-1.service 4:start:/etc/systemd/system/unit-1.service 4:start:/etc/systemd/system/unit-2.service]", fmt.Sprint(executor.Plan(nil, left)))
	})
	t.Run("destroy pod", func(t *testing.T) {
		assert.Equal(t, "[0:stop:/etc/systemd/system/pod-pod-1.service 0:stop:/etc/systemd/system/unit-1.service 0:stop:/etc/systemd/system/unit-2.service 1:remove:/etc/systemd/system/pod-pod-1.service 1:remove:/etc/systemd/system/unit-1.service 1:remove:/etc/systemd/system/unit-2.service]", fmt.Sprint(executor.Plan(left, nil)))
	})
	t.Run("change prefix", func(t *testing.T) {
		right := &allocation.Allocation{
			AllocationHeader: &allocation.AllocationHeader{
				Name:      "pod-1",
				AgentMark: 123,
				PodMark:   456,
				Namespace: "private",
			},
			AllocationFile: &allocation.AllocationFile{
				Path:   "/etc/systemd/system/pod-local-pod-1.service",
				Source: "fake",
			},
			Units: []*allocation.AllocationUnit{
				{
					AllocationFile: &allocation.AllocationFile{
						Path:   "/etc/systemd/system/unit-1.service",
						Source: "fake",
					},
					AllocationUnitHeader: &allocation.AllocationUnitHeader{
						Permanent: true,
						Transition: manifest.Transition{
							Create:  "start",
							Update:  "restart",
							Destroy: "stop",
						},
					},
				},
				{
					AllocationFile: &allocation.AllocationFile{
						Path:   "/etc/systemd/system/unit-2.service",
						Source: "fake",
					},
					AllocationUnitHeader: &allocation.AllocationUnitHeader{
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
		assert.Equal(t, "[0:stop:/etc/systemd/system/pod-pod-1.service 1:remove:/etc/systemd/system/pod-pod-1.service 2:write:/etc/systemd/system/pod-local-pod-1.service 3:enable:/etc/systemd/system/pod-local-pod-1.service 4:start:/etc/systemd/system/pod-local-pod-1.service]", fmt.Sprint(executor.Plan(left, right)))
	})
	t.Run("local to runtime", func(t *testing.T) {
		right := &allocation.Allocation{
			AllocationHeader: &allocation.AllocationHeader{
				Name:      "pod-1",
				AgentMark: 123,
				PodMark:   456,
				Namespace: "private",
			},
			AllocationFile: &allocation.AllocationFile{
				Path:   "/run/systemd/system/pod-pod-1.service",
				Source: "fake",
			},
			Units: []*allocation.AllocationUnit{
				{
					AllocationFile: &allocation.AllocationFile{
						Path:   "/run/systemd/system/unit-1.service",
						Source: "fake",
					},
					AllocationUnitHeader: &allocation.AllocationUnitHeader{
						Permanent: true,
						Transition: manifest.Transition{
							Create:  "start",
							Update:  "restart",
							Destroy: "stop",
						},
					},
				},
				{
					AllocationFile: &allocation.AllocationFile{
						Path:   "/run/systemd/system/unit-2.service",
						Source: "fake",
					},
					AllocationUnitHeader: &allocation.AllocationUnitHeader{
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
		assert.Equal(t, "[0:stop:/etc/systemd/system/pod-pod-1.service 0:stop:/etc/systemd/system/unit-1.service 0:stop:/etc/systemd/system/unit-2.service 1:remove:/etc/systemd/system/pod-pod-1.service 1:remove:/etc/systemd/system/unit-1.service 1:remove:/etc/systemd/system/unit-2.service 2:write:/run/systemd/system/pod-pod-1.service 2:write:/run/systemd/system/unit-1.service 2:write:/run/systemd/system/unit-2.service 3:enable:/run/systemd/system/pod-pod-1.service 3:enable:/run/systemd/system/unit-1.service 3:enable:/run/systemd/system/unit-2.service 4:start:/run/systemd/system/pod-pod-1.service 4:start:/run/systemd/system/unit-1.service 4:start:/run/systemd/system/unit-2.service]", fmt.Sprint(executor.Plan(left, right)))
	})
}
