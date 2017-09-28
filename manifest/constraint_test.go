// +build ide test_unit

package manifest_test

import (
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConstraint_Check(t *testing.T) {

	t.Run("equal", func(t *testing.T) {
		constraint := manifest.Constraint{
			"one,two": "${meta.field}",
		}
		assert.NoError(t, constraint.Check(map[string]string{
			"meta.field": "one,two",
		}))
	})
	t.Run("equal strict", func(t *testing.T) {
		constraint := manifest.Constraint{
			"one,two": "= ${meta.field}",
		}
		assert.NoError(t, constraint.Check(map[string]string{
			"meta.field": "one,two",
		}))
	})
	t.Run("equal strict empty", func(t *testing.T) {
		constraint := manifest.Constraint{
			"": "= ${meta.field}",
		}
		assert.NoError(t, constraint.Check(map[string]string{
			"meta.field": "",
		}))
	})
	t.Run("not equal", func(t *testing.T) {
		constraint := manifest.Constraint{
			"one,two": "${meta.field}",
		}
		assert.NoError(t, constraint.Check(map[string]string{
			"meta.field": "!= two",
		}))
	})
	t.Run("not equal empty", func(t *testing.T) {
		constraint := manifest.Constraint{
			"one,two": "${meta.field}",
		}
		assert.NoError(t, constraint.Check(map[string]string{
			"meta.field": "!= ",
		}))
	})
	t.Run("in ok", func(t *testing.T) {
		constraint := manifest.Constraint{
			"one,two": "~ ${meta.field}",
		}
		assert.NoError(t, constraint.Check(map[string]string{
			"meta.field": "one,two,three",
		}))
	})
	t.Run("in fail", func(t *testing.T) {
		constraint := manifest.Constraint{
			"one,two": "~ ${meta.field}",
		}
		assert.Error(t, constraint.Check(map[string]string{
			"meta.field": "one,three",
		}))
	})
	t.Run("not in ok", func(t *testing.T) {
		constraint := manifest.Constraint{
			"none": "!~ ${meta.field}",
		}
		assert.NoError(t, constraint.Check(map[string]string{
			"meta.field": "one,two,three",
		}))
	})
	t.Run("not in fail", func(t *testing.T) {
		constraint := manifest.Constraint{
			"one,two": "!~ ${meta.field}",
		}
		assert.Error(t, constraint.Check(map[string]string{
			"meta.field": "one,two,three",
		}))
	})
	t.Run("less ok", func(t *testing.T) {
		constraint := manifest.Constraint{
			"2": "< ${meta.num}",
		}
		assert.NoError(t, constraint.Check(map[string]string{
			"meta.num": "11",
		}))
	})
	t.Run("less fail", func(t *testing.T) {
		constraint := manifest.Constraint{
			"2": "< ${meta.num}",
		}
		assert.Error(t, constraint.Check(map[string]string{
			"meta.num": "1",
		}))
	})
	t.Run("greater ok", func(t *testing.T) {
		constraint := manifest.Constraint{
			"2": "> ${meta.num}",
		}
		assert.NoError(t, constraint.Check(map[string]string{
			"meta.num": "1",
		}))
	})
	t.Run("greater fail", func(t *testing.T) {
		constraint := manifest.Constraint{
			"2": "> ${meta.num}",
		}
		assert.Error(t, constraint.Check(map[string]string{
			"meta.num": "3",
		}))
	})
	t.Run("empty", func(t *testing.T) {
		constraint := manifest.Constraint{}
		assert.NoError(t, constraint.Check(map[string]string{
			"meta.num": "3",
		}))
	})
}

func TestConstraint_Ignore(t *testing.T) {
	constraint := manifest.Constraint{
		"${meta.a}":                       "true",
		"${resource.counter.a.allocated}": "true",
		"${resource.port.8080.allocated}": "true",
	}
	t.Run("none", func(t *testing.T) {
		res := constraint.Ignore("none.found")
		assert.Equal(t, res, manifest.Constraint{
			"${meta.a}":                       "true",
			"${resource.counter.a.allocated}": "true",
			"${resource.port.8080.allocated}": "true",
		})
	})
	t.Run("all resource", func(t *testing.T) {
		res := constraint.Ignore("resource.counter.a.allocated", "resource.port.8080.allocated")
		assert.Equal(t, res, manifest.Constraint{
			"${meta.a}": "true",
		})

	})
}
