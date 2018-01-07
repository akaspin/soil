package allocation

import (
	"strings"
)

const specVersionPrefix = "### SOIL "

// GetVersion searches for "### SOIL <version>" and returns spec version
func GetVersion(src string) (v string) {
	for _, line := range strings.Split(src, "\n") {
		if strings.HasPrefix(line, specVersionPrefix) {
			v = strings.TrimSpace(strings.TrimPrefix(line, specVersionPrefix))
			return
		}
	}
	return
}
