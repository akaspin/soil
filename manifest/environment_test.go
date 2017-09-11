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
