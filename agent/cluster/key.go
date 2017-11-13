package cluster

import (
	"path"
	"strings"
)

func NormalizeKey(v ...string) (res string) {
	res = strings.Trim(path.Clean(path.Join(v...)), "/.")
	return
}

func TrimKeyPrefix(prefix string, key string) (res string) {
	res = NormalizeKey(strings.TrimPrefix(key, prefix))
	return
}
