package fixture

import (
	"fmt"
	"github.com/eapache/go-resiliency/retrier"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func PollEqual(t *testing.T, retry int, timeout time.Duration, fn func() (res interface{}, err error), expect interface{}, doneCh chan error) {
	t.Helper()
	r := retrier.New(retrier.ConstantBackoff(retry, timeout), &retrier.DefaultClassifier{})

	retryErr := r.Run(func() (err error) {
		t.Helper()
		result, err := fn()
		if err != nil {
			return
		}
		if !assert.ObjectsAreEqual(expect, result) {
			msg1, msg2 := formatUnequalValues(expect, result)
			err = fmt.Errorf("not equal: %s (expected) != %s (actual)", msg1, msg2)
		}
		return
	})
	if retryErr != nil {
		t.Error(retryErr)
		t.Fail()
	}
	if doneCh != nil {
		select {
		case doneCh <- retryErr:
		default:
		}
	}
}

func formatUnequalValues(expected, actual interface{}) (e string, a string) {
	aType := reflect.TypeOf(expected)
	bType := reflect.TypeOf(actual)

	if aType != bType && isNumericType(aType) && isNumericType(bType) {
		return fmt.Sprintf("%v(%#v)", aType, expected),
			fmt.Sprintf("%v(%#v)", bType, actual)
	}

	return fmt.Sprintf("%#v", expected),
		fmt.Sprintf("%#v", actual)
}

func isNumericType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	}

	return false
}
