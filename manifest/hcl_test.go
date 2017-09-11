// +build ide test_unit

package manifest_test

import (
	"github.com/hashicorp/hcl"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHCL(t *testing.T) {
	type Inherited struct {
		First  string
		Second string
	}
	type Named struct {
		Name string
		Val  string
	}
	type Prey struct {
		PFirst     string
		PSecond    string
		PThird     string
		InheritedS Inherited
		InheritedP *Inherited
		Inherited  `hcl:",squash"`
		N1         []*Named
		N2         map[string]*Named
	}
	src := `
		pfirst = "overrided"
		pthird = ""
		inheriteds {
			first = "overrided"
		}
		inheritedp {
			first = "overrided"
		}
		first = "overrided"
		n1 {
			val = "v1"
		}
		n2 "a" {
			val = "a1"
		}
	`
	p := &Prey{
		PFirst:  "default",
		PSecond: "default",
		PThird:  "default",
		InheritedS: Inherited{
			Second: "default",
		},
		Inherited: Inherited{
			Second: "default",
		},
	}
	err := hcl.Decode(p, src)
	assert.NoError(t, err)
	assert.Equal(t, &Prey{
		PFirst:  "overrided",
		PSecond: "default",
		PThird:  "",
		InheritedS: Inherited{
			First:  "overrided",
			Second: "default",
		},
		InheritedP: &Inherited{
			First: "overrided",
		},
		Inherited: Inherited{
			First:  "overrided",
			Second: "default",
		},
		N1: []*Named{
			{
				Val: "v1",
			},
		},
		N2: map[string]*Named{
			"a": {
				Val: "a1",
			},
		},
	}, p)
}
