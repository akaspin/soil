package supervisor_test

import (
	"errors"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAppendError(t *testing.T) {
	t.Run("nil nil", func(t *testing.T) {
		err := supervisor.AppendError(nil, nil)
		assert.NoError(t, err)
	})

	t.Run("error nil", func(t *testing.T) {
		err1 := errors.New("1")
		err := supervisor.AppendError(err1, nil)
		assert.Equal(t, err1, err)
	})
	t.Run("nil error", func(t *testing.T) {
		err1 := errors.New("1")
		err := supervisor.AppendError(nil, err1)
		assert.Equal(t, err1, err)
	})
	t.Run("error error", func(t *testing.T) {
		err1 := errors.New("1")
		err2 := errors.New("2")
		err := supervisor.AppendError(err1, err2)
		assert.Equal(t, supervisor.MultiError{err1, err2}, err)
		assert.EqualError(t, err, "1,2")
	})
	t.Run("MultiError error", func(t *testing.T) {
		err0 := errors.New("1")
		err1 := supervisor.MultiError{err0}
		err2 := errors.New("2")
		err := supervisor.AppendError(err1, err2)
		assert.Equal(t, supervisor.MultiError{err0, err2}, err)
		assert.EqualError(t, err, "1,2")
	})
	t.Run("error MultiError", func(t *testing.T) {
		err1 := errors.New("1")
		err0 := errors.New("2")
		err2 := supervisor.MultiError{err0}
		err := supervisor.AppendError(err1, err2)
		assert.Equal(t, supervisor.MultiError{err1, err0}, err)
		assert.EqualError(t, err, "1,2")
	})
	t.Run("MultiError MultiError", func(t *testing.T) {
		err1 := errors.New("1")
		err2 := errors.New("2")
		err1m := supervisor.MultiError{err1}
		err2m := supervisor.MultiError{err2}
		err := supervisor.AppendError(err1m, err2m)
		assert.Equal(t, supervisor.MultiError{err1, err2}, err)
		assert.EqualError(t, err, "1,2")
	})
}
