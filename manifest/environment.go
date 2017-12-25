package manifest

import (
	"encoding/json"
	"regexp"
	"strings"
)

const hiddenPrefix = "__"

var (
	envRe = regexp.MustCompile(`\$\{[a-zA-Z0-9_/\-.|]+}`)
)

// FlatMap
type FlatMap map[string]string

// Merge values with environment
func (e FlatMap) Merge(env ...FlatMap) (res FlatMap) {
	res = FlatMap{}
	for _, e1 := range append([]FlatMap{e}, env...) {
		for k, v := range e1 {
			res[k] = v
		}
	}
	return
}

// Return flatmap without regex
func (e FlatMap) Filter(r ...*regexp.Regexp) (res FlatMap) {
	res = FlatMap{}
LOOP:
	for k, v := range e {
		for _, r1 := range r {
			if r1.MatchString(k) {
				continue LOOP
			}
			res[k] = v
		}
	}
	return
}

func (e FlatMap) WithJSON(key string) (j FlatMap) {
	buf, _ := json.Marshal(e)
	j = e.Merge(FlatMap{
		key: string(buf),
	})
	return
}

// Interpolate source
func (e FlatMap) Interpolate(source string) (res string) {
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
