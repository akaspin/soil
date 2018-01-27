package cluster

import (
	"path"
	"strings"
)

func NormalizeKey(v ...string) (res string) {
	return strings.Trim(path.Clean(path.Join(v...)), "/.")
}

func TrimKeyPrefix(prefix string, key string) (res string) {
	return NormalizeKey(strings.TrimPrefix(key, prefix))
}
