// +build ide test_unit

package provider_test

import (
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/provider"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPlan(t *testing.T) {
	left := allocation.ProviderSlice{
		{
			Kind: "1",
			Name: "1",
		},
		{
			Kind: "1",
			Name: "2",
			Config: map[string]interface{}{
				"a": 1,
			},
		},
	}
	right := allocation.ProviderSlice{
		{
			Kind: "2",
			Name: "2",
			Config: map[string]interface{}{
				"a": 2,
			},
		},
		{
			Kind: "test",
			Name: "3",
		},
	}

	t.Run(`create all`, func(t *testing.T) {
		c, u, d := provider.Plan(nil, left)
		assert.Equal(t, left, c)
		assert.Empty(t, u)
		assert.Empty(t, d)
	})
	t.Run(`destroy all`, func(t *testing.T) {
		c, u, d := provider.Plan(left, nil)
		assert.Empty(t, c)
		assert.Empty(t, u)
		assert.Equal(t, []string{"1", "2"}, d)
	})
	t.Run(`change c3 u2 d1`, func(t *testing.T) {
		c, u, d := provider.Plan(left, right)
		assert.Equal(t, allocation.ProviderSlice{
			{
				Kind: "test",
				Name: "3",
			},
		}, c)
		assert.Equal(t, allocation.ProviderSlice{
			&allocation.Provider{
				Kind: "2",
				Name: "2",
				Config: map[string]interface{}{
					"a": int(2),
				},
			},
		}, u)
		assert.Equal(t, []string{"1"}, d)
	})
	t.Run(`change 2`, func(t *testing.T) {
		r1 := allocation.ProviderSlice{
			{
				Kind: "1",
				Name: "1",
			},
			{
				Kind: "1",
				Name: "2",
				Config: map[string]interface{}{
					"a": 2,
				},
			},
		}
		c, u, d := provider.Plan(left, r1)
		assert.Empty(t, c)
		assert.Equal(t, allocation.ProviderSlice{
			{
				Kind: "1",
				Name: "2",
				Config: map[string]interface{}{
					"a": 2,
				},
			},
		}, u)
		assert.Empty(t, d)
	})

}
