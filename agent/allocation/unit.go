package allocation

import (
	"encoding/json"
	"fmt"
	"github.com/akaspin/soil/manifest"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
)

const (
	unitSpecPrefix = "### UNIT "
	unitV2Prefix   = "### UNIT_V2 "
)

type UnitSlice []*Unit

func (s *UnitSlice) GetEmpty(paths SystemPaths) (empty ItemUnmarshaller) {
	empty = &Unit{
		UnitFile: UnitFile{
			SystemPaths: paths,
		},
	}
	return
}

func (s *UnitSlice) GetVersionPrefix(v string) (p string) {
	p = unitSpecPrefix
	return
}

func (s *UnitSlice) AppendItem(v ItemUnmarshaller) {
	*s = append(*s, v.(*Unit))
}

type Unit struct {
	UnitFile
	manifest.Transition `json:",squash"`
}

func (u *Unit) MarshalLine(w io.Writer) (err error) {
	if _, err = w.Write([]byte(unitSpecPrefix)); err != nil {
		return
	}
	err = json.NewEncoder(w).Encode(u)
	return
}

// UnmarshalItem parses one line from manifest
func (u *Unit) UnmarshalItem(line string, spec SpecMeta, paths SystemPaths) (err error) {
	u.SystemPaths = paths
	switch spec.Revision {
	case "":
		// v1
		if _, err = fmt.Sscanf(line, "### UNIT %s ", &u.UnitFile.Path); err != nil {
			return
		}
		line = strings.TrimPrefix(line, fmt.Sprintf("%s%s ", unitSpecPrefix, u.UnitFile.Path))
		if err = json.NewDecoder(strings.NewReader(line)).Decode(u); err != nil {
			return
		}
	case SpecRevision:
		// v2
		if err = json.NewDecoder(strings.NewReader(strings.TrimPrefix(line, unitSpecPrefix))).Decode(u); err != nil {
			return
		}
	}
	src, err := ioutil.ReadFile(u.UnitFile.Path)
	if err != nil {
		return
	}
	u.UnitFile.Source = string(src)
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
