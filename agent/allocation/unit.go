package allocation

import (
	"encoding/json"
	"fmt"
	"github.com/akaspin/soil/manifest"
	"io"
	"io/ioutil"
	"path/filepath"
)

type Unit struct {
	*UnitFile
	manifest.Transition `json:",squash"`
}

func (u *Unit) MarshalHeader(w io.Writer, encoder *json.Encoder) (err error) {
	if _, err = fmt.Fprintf(w, "### UNIT %s ", u.UnitFile.Path); err != nil {
		return
	}
	err = encoder.Encode(&u.Transition)
	return
}

type UnitFile struct {
	SystemPaths SystemPaths
	Path        string
	Source      string
}

func NewUnitFile(unitName string, paths SystemPaths, runtime bool) (f *UnitFile) {
	basePath := paths.Local
	if runtime {
		basePath = paths.Runtime
	}
	f = &UnitFile{
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
