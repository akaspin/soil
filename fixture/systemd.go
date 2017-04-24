package fixture

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/coreos/go-systemd/dbus"
	"github.com/coreos/go-systemd/unit"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Systemd struct {
	Dir    string
	Prefix string
}

func NewSystemd(dir, prefix string) *Systemd {
	return &Systemd{
		Dir:    dir,
		Prefix: prefix,
	}
}

func (s *Systemd) DeployPod(name string, n int) (err error) {
	isRuntime := strings.HasPrefix(s.Dir, "/run")
	podUnitName := fmt.Sprintf("%s-%s.service", s.Prefix, name)

	var unitNames []string
	podHeaderJ := map[string]interface{}{
		"PodMark":   123,
		"AgentMark": 456,
		"Namespace": "private",
	}

	var unitS []string
	for i := 0; i < n; i++ {
		unitName := fmt.Sprintf("%s-%d.service", name, i)
		unitNames = append(unitNames, unitName)
		unitHeaderJ, _ := json.Marshal(map[string]interface{}{
			"Create":    "start",
			"Update":    "restart",
			"Destroy":   "stop",
			"Permanent": true,
		})
		unitS = append(unitS, fmt.Sprintf("### UNIT %s %s", filepath.Join(s.Dir, unitName), string(unitHeaderJ)))
		unitSrc := fmt.Sprintf(`[Unit]
Description=Unit %s
[Service]
ExecStart=/usr/bin/sleep inf
[Install]
WantedBy=multi-user.target
`, unitName)
		if err = ioutil.WriteFile(filepath.Join(s.Dir, unitName), []byte(unitSrc), 0775); err != nil {
			return
		}
	}

	// POD
	headerJSON, err := json.Marshal(podHeaderJ)
	if err != nil {
		return
	}
	podSrc := fmt.Sprintf(`### POD %s %s
%s
[Unit]
Description=%s
Before=%s
[Service]
ExecStart=/usr/bin/sleep inf
[Install]
WantedBy=multi-user.target
`, name, string(headerJSON), strings.Join(unitS, "\n"), name, strings.Join(unitNames, " "))
	if err = ioutil.WriteFile(filepath.Join(s.Dir, podUnitName), []byte(podSrc), 755); err != nil {
		return
	}
	conn, err := dbus.New()
	if err != nil {
		return
	}
	defer conn.Close()

	if err = conn.Reload(); err != nil {
		return
	}
	if _, _, err = conn.EnableUnitFiles(append(unitNames, podUnitName), isRuntime, false); err != nil {
		return
	}
	for _, n := range append([]string{podUnitName}, unitNames...) {
		ch := make(chan string)
		if _, err = conn.StartUnit(n, "replace", ch); err != nil {
			return
		}
		<-ch
	}
	return
}

func (s *Systemd) DestroyPod(name ...string) (err error) {
	var unitNames []string
	for _, n := range name {
		unitNames = append(unitNames, fmt.Sprintf("%s-%s.service", s.Prefix, n))
	}
	conn, err := dbus.New()
	if err != nil {
		return
	}
	defer conn.Close()
	fs, err := conn.ListUnitFilesByPatterns([]string{}, unitNames)
	if err != nil {
		return
	}
	for _, f := range fs {
		body, readErr := ioutil.ReadFile(f.Path)
		if readErr != nil {
			continue
		}
		if strings.HasPrefix(string(body), "### POD ") {
			if err = s.destroyPod(conn, f.Path, body); err != nil {
				fmt.Printf("ERR can't destroy pod %s", f.Path)
				continue
			}
		}
	}
	return
}

func (s *Systemd) Cleanup() (err error) {
	err = s.DestroyPod("*")
	return
}

func (s *Systemd) ListPods() (res map[string]string, err error) {
	conn, err := dbus.New()
	if err != nil {
		return
	}
	defer conn.Close()

	fs, err := conn.ListUnitFilesByPatterns([]string{}, []string{s.Prefix + "-*.service"})
	if err != nil {
		return
	}
	res = map[string]string{}
	for _, f := range fs {
		body, readErr := ioutil.ReadFile(f.Path)
		if readErr != nil {
			continue
		}
		var name string
		var parseErr error
		if _, parseErr = fmt.Sscanf(string(body), "### POD %s ", &name); parseErr != nil {
			continue
		}
		res[name] = f.Path
	}
	return
}

func (s *Systemd) destroyPod(conn *dbus.Conn, path string, src []byte) (err error) {
	isRuntime := strings.HasPrefix(path, "/run")
	unitSpec, err := unit.Deserialize(bytes.NewReader(src))
	if err != nil {
		return
	}
	unitNames := []string{filepath.Base(path)}
	for _, prop := range unitSpec {
		if prop.Name == "Before" && prop.Section == "Unit" {
			unitNames = append(unitNames, strings.Split(prop.Value, " ")...)
		}
	}

	conn.DisableUnitFiles(unitNames, isRuntime)
	for _, u := range unitNames {
		// ch := make(chan string)
		conn.StopUnit(u, "replace", nil)
		//<-ch
	}

	files1, err := conn.ListUnitFilesByPatterns([]string{}, unitNames)
	if err != nil {
		return
	}
	for _, f := range files1 {
		os.Remove(f.Path)
	}
	conn.Reload()

	return
}
