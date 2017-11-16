package fixture

import (
	"fmt"
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

func RandomPorts(t *testing.T, count int) (res []int) {
	t.Helper()
	var errs []error
	var listeners []net.Listener
	for i := 0; i < count; i++ {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Error(err)
			continue
		}
		listeners = append(listeners, l)
	}
	defer func() {
		for _, listener := range listeners {
			listener.Close()
		}
	}()
	if len(errs) > 0 {
		t.Log(fmt.Errorf("%v", errs))
		t.FailNow()
	}
	for _, listener := range listeners {
		res = append(res, listener.Addr().(*net.TCPAddr).Port)
	}
	return
}

func GetLocalIP(t *testing.T) string {
	t.Helper()
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		t.Error(err)
		t.FailNow()
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	t.Error("can't find local IP")
	t.FailNow()
	return ""
}
