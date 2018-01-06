package allocation

import (
	"encoding/json"
	"fmt"
	"github.com/akaspin/soil/manifest"
	"io"
	"io/ioutil"
	"path/filepath"
)

const (
	unitV1Prefix = "### BLOB "
	unitV2Prefix = "### BLOB_V2 "
)

type UnitSlice []*Unit

type Unit struct {
	UnitFile
	manifest.Transition `json:",squash"`
}

func (u *Unit) MarshalLine(w io.Writer) (err error) {
	if _, err = w.Write([]byte(unitV2Prefix)); err != nil {
		return
	}
	err = json.NewEncoder(w).Encode(u)
	return
}

// UnmarshalItem parses one line from manifest
//
//	  v1: ### UNIT ...
//	  v2: ### UNIT_V2 ...
func (u *Unit) UnmarshalItem(line string) (err error) {
	return
}

func (u *Unit) MarshalHeader(w io.Writer, encoder *json.Encoder) (err error) {
	if _, err = fmt.Fprintf(w, "### UNIT %s ", u.UnitFile.Path); err != nil {
		return
	}
	err = encoder.Encode(&u.Transition)
	return
}

type UnitFile struct {
	SystemPaths SystemPaths `json:"-"`
	Path        string
	Source      string `json:"-"`
}

func NewUnitFile(unitName string, paths SystemPaths, runtime bool) (f UnitFile) {
	basePath := paths.Local
	if runtime {
		basePath = paths.Runtime
	}
	f = UnitFile{
		SystemPaths: paths,
		Path:        filepath.Join(basePath, unitName),
	}
	return
}

func (f *UnitFile) Read() (err error) {
	src, err := ioutil.ReadFile(f.Path)
	if err != nil {
		return
	}
	f.Source = string(src)
	return
}

func (f *UnitFile) Write() (err error) {
	err = ioutil.WriteFile(f.Path, []byte(f.Source), 755)
	return
}

func (f *UnitFile) UnitName() (res string) {
	res = filepath.Base(f.Path)
	return
}

func (f *UnitFile) IsRuntime() (res bool) {
	res = filepath.Dir(f.Path) == f.SystemPaths.Runtime
	return
}
