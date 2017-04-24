package cut_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

type sieve interface {
	fn()
}

type i1 struct {}
func (i *i1) fn() {}

type i2 struct {}
func (i *i2) fn() {}

type ci2 struct {
	*i1
	*i2
}

type ci1 struct {
	*i1
}

func TestCompose_CI1(t *testing.T) {
	c1 := &ci1{
		&i1{},
	}
	_, ok := (interface{})(c1).(sieve)
	assert.True(t, ok)
}

func TestCompose_CI2(t *testing.T) {
	c1 := &ci2{
		&i1{},
		&i2{},
	}
	_, ok := (interface{})(c1).(sieve)
	assert.False(t, ok)
}
