// +build ide test_unit

package provision_test

import (
	"github.com/akaspin/soil/agent/provision"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEvaluation_Plan(t *testing.T) {
	left1 := makeAllocations(t, "testdata/evaluation_test_left.hcl")[0]

	t.Run("0 - noop pod", func(t *testing.T) {
		right := makeAllocations(t, "testdata/evaluation_test_left.hcl")[0]
		evaluation := provision.NewEvaluation(left1, right)
		assert.Equal(t, "[]", evaluation.Explain())
	})
	t.Run("1 - unit-1 perm to disabled", func(t *testing.T) {
		right := makeAllocations(t, "testdata/evaluation_test_1_right.hcl")[0]
		evaluation := provision.NewEvaluation(left1, right)
		assert.Equal(t, "[2:write-unit:/etc/systemd/system/pod-private-pod-1.service 3:disable-unit:/etc/systemd/system/unit-1.service 3:enable-unit:/etc/systemd/system/pod-private-pod-1.service 4:restart:/etc/systemd/system/pod-private-pod-1.service]", evaluation.Explain())
	})
	t.Run("2 - update unit-1 and file", func(t *testing.T) {
		right := makeAllocations(t, "testdata/evaluation_test_2_right.hcl")[0]
		evaluation := provision.NewEvaluation(left1, right)
		assert.Equal(t, "[2:write-blob:/etc/test1 2:write-unit:/etc/systemd/system/pod-private-pod-1.service 2:write-unit:/etc/systemd/system/unit-1.service 3:enable-unit:/etc/systemd/system/pod-private-pod-1.service 3:enable-unit:/etc/systemd/system/unit-1.service 4:restart:/etc/systemd/system/pod-private-pod-1.service 4:restart:/etc/systemd/system/unit-1.service]", evaluation.Explain())
	})
	t.Run("3 - create pod form left", func(t *testing.T) {
		evaluation := provision.NewEvaluation(nil, left1)
		assert.Equal(t, "[2:write-blob:/etc/test1 2:write-unit:/etc/systemd/system/pod-private-pod-1.service 2:write-unit:/etc/systemd/system/unit-1.service 2:write-unit:/etc/systemd/system/unit-2.service 3:enable-unit:/etc/systemd/system/pod-private-pod-1.service 3:enable-unit:/etc/systemd/system/unit-1.service 3:enable-unit:/etc/systemd/system/unit-2.service 4:start:/etc/systemd/system/pod-private-pod-1.service 4:start:/etc/systemd/system/unit-1.service 4:start:/etc/systemd/system/unit-2.service]", evaluation.Explain())
	})
	t.Run("4 - destroy pod", func(t *testing.T) {
		evaluation := provision.NewEvaluation(left1, nil)
		assert.Equal(t, "[0:stop:/etc/systemd/system/pod-private-pod-1.service 0:stop:/etc/systemd/system/unit-1.service 0:stop:/etc/systemd/system/unit-2.service 1:delete-unit:/etc/systemd/system/pod-private-pod-1.service 1:delete-unit:/etc/systemd/system/unit-1.service 1:delete-unit:/etc/systemd/system/unit-2.service 5:delete-blob:/etc/test1]", evaluation.Explain())
	})
	t.Run("5 - local to runtime", func(t *testing.T) {
		right := makeAllocations(t, "testdata/evaluation_test_5_right.hcl")[0]
		evaluation := provision.NewEvaluation(left1, right)
		assert.Equal(t, "[0:stop:/etc/systemd/system/pod-private-pod-1.service 0:stop:/etc/systemd/system/unit-1.service 0:stop:/etc/systemd/system/unit-2.service 1:delete-unit:/etc/systemd/system/pod-private-pod-1.service 1:delete-unit:/etc/systemd/system/unit-1.service 1:delete-unit:/etc/systemd/system/unit-2.service 2:write-unit:/run/systemd/system/pod-private-pod-1.service 2:write-unit:/run/systemd/system/unit-1.service 2:write-unit:/run/systemd/system/unit-2.service 3:enable-unit:/run/systemd/system/pod-private-pod-1.service 3:enable-unit:/run/systemd/system/unit-1.service 3:enable-unit:/run/systemd/system/unit-2.service 4:start:/run/systemd/system/pod-private-pod-1.service 4:start:/run/systemd/system/unit-1.service 4:start:/run/systemd/system/unit-2.service]", evaluation.Explain())
	})
}
