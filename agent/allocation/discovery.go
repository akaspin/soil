package allocation

import "github.com/coreos/go-systemd/dbus"

const DefaultPodPrefix = "pod-*"

func dbusDiscoveryFunc(prefix ...string) (res []string, err error) {
	conn, err := dbus.New()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	files, err := conn.ListUnitFilesByPatterns([]string{}, prefix)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		res = append(res, f.Path)
	}
	return res, nil
}

func DefaultDbusDiscoveryFunc() (res []string, err error) {
	return dbusDiscoveryFunc(DefaultPodPrefix)
}

func GetZeroDiscoveryFunc(paths ...string) func() ([]string, error) {
	return func() ([]string, error) {
		return paths, nil
	}
}
