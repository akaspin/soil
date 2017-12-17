package manifest

import (
	"regexp"
	"strings"
)

const hiddenPrefix = "__"

var (
	envRe = regexp.MustCompile(`\$\{[a-zA-Z0-9_/\-.|]+}`)
)

// Environment
type Environment map[string]string

// Merge values with environment
func (e Environment) Merge(env ...Environment) (res Environment) {
	res = Environment{}
	for _, e1 := range append([]Environment{e}, env...) {
		for k, v := range e1 {
			res[k] = v
		}
	}
	return
}

// Interpolate source
func (e Environment) Interpolate(source string) (res string) {
	res = envRe.ReplaceAllStringFunc(source, func(arg string) string {
		var hasDefaultValue bool
		var defaultValue string
		stripped := arg[2 : len(arg)-1]
		split := strings.SplitN(stripped, "|", 2)
		if len(split) == 2 {
			hasDefaultValue = true
			defaultValue = split[1]
			stripped = split[0]
		}
		if value, ok := e[stripped]; ok {
			return value
		}
		if hasDefaultValue {
			return defaultValue
		}
		return arg
	})
	return
}

func ExtractEnv(v string) (res []string) {
	res1 := envRe.FindAllString(v, -1)
	for _, r := range res1 {
		res = append(res, r[2:len(r)-1])
	}
	return
}

func Interpolate(v string, env ...map[string]string) (res string) {
	res = envRe.ReplaceAllStringFunc(v, func(arg string) string {
		var hasDefaultValue bool
		var defaultValue string
		stripped := arg[2 : len(arg)-1]
		split := strings.SplitN(stripped, "|", 2)
		if len(split) == 2 {
			hasDefaultValue = true
			defaultValue = split[1]
			stripped = split[0]
		}
		for _, envChunk := range env {
			if value, ok := envChunk[stripped]; ok {
				return value
			}
		}
		if hasDefaultValue {
			return defaultValue
		}
		return arg
	})
	return
}
