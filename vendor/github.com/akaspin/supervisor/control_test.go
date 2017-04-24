package supervisor_test

import (
	"testing"
	"github.com/akaspin/supervisor"
	"golang.org/x/net/context"
	"time"
	"github.com/stretchr/testify/assert"
)

func TestNewControlTimeout(t *testing.T) {
	ctl := supervisor.NewControlTimeout(context.TODO(), time.Millisecond * 100)
	ctl.Open()

	ctl.Acquire()

	ctl.Close()
	err := ctl.Wait()
	assert.Error(t, err)
	assert.Equal(t, err, supervisor.CloseTimeoutExceeded)
}
