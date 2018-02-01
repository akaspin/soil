package fixture

import (
	"bytes"
	"fmt"
	"github.com/coreos/go-systemd/dbus"
	"github.com/mitchellh/hashstructure"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
)

func WriteTemplate(w io.Writer, source string, env map[string]interface{}) (err error) {
	var lines []string
	for _, line := range strings.Split(source, "\n") {
		lines = append(lines, strings.TrimSpace(line))
	}
	tpl, err := template.New("T").Parse(strings.Join(lines, "\n"))
	if err != nil {
		return err
	}
	return tpl.Execute(w, env)
}

// CreateUnit creates unit with given body template
func CreateUnit(path, source string, env map[string]interface{}) (err error) {
	conn, err := dbus.New()
	if err != nil {
		return err
	}
	defer conn.Close()

	if err = func() error {
		f, err1 := os.Create(path)
		if err1 != nil {
			return err1
		}
		defer f.Close()
		if err1 = WriteTemplate(f, source, env); err != nil {
			return err1
		}
		return nil
	}(); err != nil {
		return err
	}

	if err = conn.Reload(); err != nil {
		return err
	}
	ch := make(chan string, 1)
	_, err = conn.StartUnit(filepath.Base(path), "replace", ch)
	<-ch
	return
}

// CheckUnitBody compares file contents with template
func CheckUnitBody(path, source string, env map[string]interface{}) (err error) {
	var buf bytes.Buffer
	if err = WriteTemplate(&buf, source, env); err != nil {
		return
	}
	data, err := ioutil.ReadFile(path)
	if buf.String() != string(data) {
		return fmt.Errorf("(expect)%s != (actual)%s", buf.String(), string(data))
	}
	return nil
}

func CheckUnitHashes(names []string, states map[string]uint64) (err error) {
	conn, err := dbus.New()
	if err != nil {
		return err
	}
	defer conn.Close()
	l, err := conn.ListUnitFilesByPatterns([]string{}, names)
	if err != nil {
		return err
	}
	res := map[string]uint64{}
	for _, u := range l {
		var data []byte
		if data, err = ioutil.ReadFile(u.Path); err != nil {
			return err
		}
		//println(string(data))
		res[u.Path], _ = hashstructure.Hash(data, nil)
	}
	if !reflect.DeepEqual(states, res) {
		return fmt.Errorf("not equal (expected)%#v != (actual)%#v", states, res)
	}
	return nil
}

// UnitStatesFn returns function to check all states for units founded
// by patterns. Expect should be "unit.service":"systemd-state"
func UnitStatesFn(patterns []string, expect map[string]string) func() error {
	return func() error {
		conn, err := dbus.New()
		if err != nil {
			return err
		}
		defer conn.Close()
		l, err := conn.ListUnitsByPatterns([]string{}, patterns)
		if err != nil {
			return err
		}
		res := map[string]string{}
		for _, u := range l {
			res[u.Name] = u.ActiveState
		}
		if !reflect.DeepEqual(expect, res) {
			return fmt.Errorf("not equal (expected)%#v != (actual)%#v", expect, res)
		}
		return nil
	}
}

// DestroyUnits disables and destroys units with given patterns
func DestroyUnits(patterns ...string) (err error) {
	conn, err := dbus.New()
	if err != nil {
		return err
	}
	defer conn.Close()
	unitFiles, err := conn.ListUnitFilesByPatterns([]string{}, patterns)

	//pretty.Log(unitFiles)
	if err != nil {
		return err
	}
	for _, u := range unitFiles {
		conn.DisableUnitFiles([]string{filepath.Base(u.Path)}, strings.HasPrefix(u.Path, "/run"))
		conn.StopUnit(filepath.Base(u.Path), "replace", nil)
		os.Remove(u.Path)
	}
	conn.Reload()
	return nil
}
