package fixture

import (
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

type WaitConfig struct {
	Retry time.Duration
	Retries int
	Timeout time.Duration
}

func DefaultWaitConfig() (c WaitConfig) {
	c= WaitConfig{
		Retry: time.Millisecond * 100,
		Retries: 100,
		Timeout: time.Minute,
	}
	return
}
