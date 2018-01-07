// +build ide test_unit

package allocation_test

import (
	"fmt"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestGetVersion(t *testing.T) {
	t.Run(`none`, func(t *testing.T) {
		assert.Equal(t, "", allocation.GetVersion("### POD \n"))
	})
	for i := 1; i < 10; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, strconv.Itoa(i), allocation.GetVersion(fmt.Sprintf(
				"### SOIL %d\n### POD ", i,
			)))
		})
	}
}
