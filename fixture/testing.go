package fixture

import (
	"strings"
	"testing"
	"time"
)

func TestName(t *testing.T) (res string) {
	t.Helper()
	return strings.ToLower(strings.Join(strings.Split(t.Name(), "/"), "__"))
}

// Poll while provided fn return no error
func WaitNoError(t *testing.T, config WaitConfig, fn func() error) {
	t.Helper()
	var err error
	var i int
	for i = 0; i < config.Retries; i++ {
		//println(fmt.Sprintf(`>>> retry %d of %d: %v`, i, config.Retries, err))
		if err = fn(); err == nil {
			break
		}
		//println(fmt.Sprintf(`<<< retry %d of %d: %v`, i, config.Retries, err))
		time.Sleep(config.Retry)
	}
	if err != nil {
		t.Errorf(`%v after %d if %d retries`, err, i, config.Retries)
		t.Fail()
	}
}

func WaitNoError10(t *testing.T, fn func() error) {
	t.Helper()
	WaitNoError(t, DefaultWaitConfig(), fn)
}

type WaitConfig struct {
	Retry   time.Duration
	Retries int
}

func DefaultWaitConfig() (c WaitConfig) {
	return WaitConfig{
		Retry:   time.Millisecond * 100,
		Retries: 100,
	}
}
