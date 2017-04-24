package agent

import (
	"fmt"
	"github.com/mitchellh/hashstructure"
	"regexp"
)

var (
	envRe = regexp.MustCompile(`\$\{[a-zA-Z0-9_\-.]+}`)
)

// Environment handler
type Environment struct {
	Fields []map[string]string
}

func NewEnvironment(envs ...map[string]string) (m *Environment) {
	m = &Environment{
		Fields: envs,
	}
	return
}

func (m *Environment) Mark() (res uint64) {
	res, _ = hashstructure.Hash(m, nil)
	return
}

// Assert
func (m *Environment) Assert(v map[string]string) (err error) {
	for k, v := range v {
		iK := m.Interpolate(k)
		iV := m.Interpolate(v)
		if iK != iV {
			err = fmt.Errorf("constraint %s:%s (%s:%s) failed", iK, iV, k, v)
			return
		}
	}
	return
}

// Interpolate template
func (m *Environment) Interpolate(v string, override ...map[string]string) (res string) {
	res = envRe.ReplaceAllStringFunc(v, func(arg string) string {
		stripped := arg[2 : len(arg)-1]
		for _, env := range append(override, m.Fields...) {
			if value, ok := env[stripped]; ok {
				return value
			}
		}
		return arg
	})
	return
}
