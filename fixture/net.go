package fixture

import (
	"net"
	"testing"
)

func RandomPort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}
