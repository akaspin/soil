package supervisor_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestTrap_Wait(t *testing.T) {
	for i := 0; i < compositeTestIterations; i++ {
		t.Run("with error "+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			trap := supervisor.NewTrap(context.Background())
			trap.Open()
			go func() {
				assert.EqualError(t, trap.Wait(), "bang")
			}()

			trap.Trap(errors.New("bang"))
			assert.EqualError(t, trap.Wait(), "bang")
		})
		t.Run("no error "+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			trap := supervisor.NewTrap(context.Background())
			trap.Open()
			go func() {
				assert.NoError(t, trap.Wait())
			}()

			trap.Close()
			assert.NoError(t, trap.Wait())
		})
	}
}

func ExampleTrap_Wait() {
	trap := supervisor.NewTrap(context.Background())
	trap.Open()
	go func() {
		if err := trap.Wait(); err != nil && err.Error() != "bang" {
			fmt.Println(trap.Wait())
		}
	}()

	trap.Trap(errors.New("bang"))
	fmt.Println(trap.Wait())
	// Output: bang
}
