package errslice_test

import (
	"errors"
	"github.com/akaspin/errslice"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAppend(t *testing.T) {
	t.Run("nil nil", func(t *testing.T) {
		err := errslice.Append(nil, nil)
		assert.NoError(t, err)
	})
	t.Run("error nil", func(t *testing.T) {
		err1 := errors.New("1")
		err := errslice.Append(err1, nil)
		assert.Equal(t, err1, err)
	})
	t.Run("nil error", func(t *testing.T) {
		err1 := errors.New("1")
		err := errslice.Append(nil, err1)
		assert.Equal(t, err1, err)
	})
	t.Run("error error", func(t *testing.T) {
		err1 := errors.New("1")
		err2 := errors.New("2")
		err := errslice.Append(err1, err2)
		assert.Equal(t, errslice.Error{err1, err2}, err)
		assert.EqualError(t, err, "1,2")
	})
	t.Run("Error error", func(t *testing.T) {
		err0 := errors.New("1")
		err1 := errslice.Error{err0}
		err2 := errors.New("2")
		err := errslice.Append(err1, err2)
		assert.Equal(t, errslice.Error{err0, err2}, err)
		assert.EqualError(t, err, "1,2")
	})
	t.Run("error Error", func(t *testing.T) {
		err1 := errors.New("1")
		err0 := errors.New("2")
		err2 := errslice.Error{err0}
		err := errslice.Append(err1, err2)
		assert.Equal(t, errslice.Error{err1, err0}, err)
		assert.EqualError(t, err, "1,2")
	})
	t.Run("Error Error", func(t *testing.T) {
		err1 := errors.New("1")
		err2 := errors.New("2")
		err1m := errslice.Error{err1}
		err2m := errslice.Error{err2}
		err := errslice.Append(err1m, err2m)
		assert.Equal(t, errslice.Error{err1, err2}, err)
		assert.EqualError(t, err, "1,2")
	})
}

