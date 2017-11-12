package fixture

import (
	"strings"
	"testing"
)

func TestName(t *testing.T) (res string) {
	t.Helper()
	res = strings.ToLower(strings.Join(strings.Split(t.Name(), "/"), "__"))
	return
}
