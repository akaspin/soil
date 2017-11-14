package fixture

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func TestName(t *testing.T) (res string) {
	t.Helper()
	res = strings.ToLower(strings.Join(strings.Split(t.Name(), "/"), "__"))
	return
}

// Poll while provided fn return no error
func WaitNoError(t *testing.T, retry time.Duration, retries int, fn func() error) {
	t.Helper()
	var err error
	for i := 0; i < retries; i++ {
		if err = fn(); err == nil {
			break
		}
		t.Logf(`retry %d of %d: %v`, i, retries, err)
	}
	assert.NoError(t, err)
}
