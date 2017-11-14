// +build ide test_unit

package cluster_test

import (
	"github.com/akaspin/soil/agent/cluster"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNormalizeKey(t *testing.T) {
	for expect, chunks := range map[string][]string{
		"":   {""},
		"a":  {"/", "a", "/"},
		"a1": {"/a1", "/"},
		"a2": {"", "a2", ""},
		"a3": {"a3"},
	} {
		assert.Equal(t, expect, cluster.NormalizeKey(chunks...))
	}
}

func TestTrimKeyPrefix(t *testing.T) {
	assert.Equal(t, "", cluster.TrimKeyPrefix("a", "a"))
	assert.Equal(t, "test/a", cluster.TrimKeyPrefix("test1", "test1/test/a"))
}
