package fixture

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestName(t *testing.T) (res string) {
	t.Helper()
	res = strings.ToLower(strings.Join(strings.Split(t.Name(), "/"), "__"))
	return
}

func WaitNoError10(fn func() error) (err error) {
	return WaitNoError(DefaultWaitConfig(), fn)
}

func WaitNoError(config WaitConfig, fn func() error) (err error) {
	var i int
	for i = 0; i < config.Retries; i++ {
		if err = fn(); err == nil {
			return nil
		}
		time.Sleep(config.Retry)
	}
	if err != nil {
		return fmt.Errorf(`%v after %d if %d retries`, err, i, config.Retries)
	}
	return nil
}

// Poll while provided fn return no error
func WaitNoErrorT(t *testing.T, config WaitConfig, fn func() error) {
	t.Helper()
	if err := WaitNoError(config, fn); err != nil {
		t.Error(err)
		t.Fail()
	}
}

func WaitNoErrorT10(t *testing.T, fn func() error) {
	t.Helper()
	WaitNoErrorT(t, DefaultWaitConfig(), fn)
}

type WaitConfig struct {
	Retry   time.Duration
	Retries int
}

func DefaultWaitConfig() (c WaitConfig) {
	c = WaitConfig{
		Retry:   time.Millisecond * 100,
		Retries: 100,
	}
	return
}
