package manifest

import (
	"encoding/json"
	"regexp"
)

const hiddenPrefix = "__"

var (
	envRe = regexp.MustCompile(`\$\{[a-zA-Z0-9_\-.]+}`)
)

func ExtractEnv(v string) (res []string) {
	res1 := envRe.FindAllString(v, -1)
	for _, r := range res1 {
		res = append(res, r[2:len(r)-1])
	}
	return
}

func Interpolate(v string, env ...map[string]string) (res string) {
	res = envRe.ReplaceAllStringFunc(v, func(arg string) string {
		stripped := arg[2 : len(arg)-1]
		for _, envChunk := range env {
			if value, ok := envChunk[stripped]; ok {
				return value
			}
		}
		return arg
	})
	return
}

func MapToJson(v map[string]string) (res string, err error) {
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	res = string(data)
	return
}
