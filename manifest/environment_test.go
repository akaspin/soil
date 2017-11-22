// +build ide test_unit

package manifest_test

import (
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExtractEnv(t *testing.T) {

	t.Run("single", func(t *testing.T) {
		res := manifest.ExtractEnv("${one.two}")
		assert.Equal(t, []string{"one.two"}, res)
	})
	t.Run("multi", func(t *testing.T) {
		res := manifest.ExtractEnv("abv${one.two}cf${one.one}")
		assert.Equal(t, []string{"one.two", "one.one"}, res)
	})
}

func TestInterpolate(t *testing.T) {
	t.Run(`ok`, func(t *testing.T) {
		assert.Equal(t, "1", manifest.Interpolate(`${test.env}`, map[string]string{
			"test.env": "1",
		}))
	})
	t.Run(`not found`, func(t *testing.T) {
		assert.Equal(t, "${test.env}", manifest.Interpolate(`${test.env}`, map[string]string{
			"test.env1": "1",
		}))
	})
	t.Run(`default not found`, func(t *testing.T) {
		assert.Equal(t, "2", manifest.Interpolate(`${test.env|2}`, map[string]string{
			"test.env1": "1",
		}))
	})
	t.Run(`default ok`, func(t *testing.T) {
		assert.Equal(t, "1", manifest.Interpolate(`${test.env|2}`, map[string]string{
			"test.env": "1",
		}))
	})
}
