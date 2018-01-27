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

func (s *UnitSlice) GetEmpty(paths SystemPaths) (empty Asset) {
	return &Unit{
		UnitFile: UnitFile{
			SystemPaths: paths,
		},
	}
}

func (s *UnitSlice) GetVersionPrefix(v string) (p string) {
	return unitSpecPrefix
}

func (s *UnitSlice) AppendItem(v Asset) {
	*s = append(*s, v.(*Unit))
}

type Unit struct {
	UnitFile
	manifest.Transition `json:",squash"`
}

func (u *Unit) MarshalSpec(w io.Writer) (err error) {
	if _, err = w.Write([]byte(unitSpecPrefix)); err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(u)
}

// UnmarshalSpec parses one line from manifest
func (u *Unit) UnmarshalSpec(line string, spec Spec, paths SystemPaths) (err error) {
	u.SystemPaths = paths
	switch spec.Revision {
	case "":
		// v1
		if _, err = fmt.Sscanf(line, "### UNIT %s ", &u.UnitFile.Path); err != nil {
			return err
		}
		line = strings.TrimPrefix(line, fmt.Sprintf("%s%s ", unitSpecPrefix, u.UnitFile.Path))
		if err = json.NewDecoder(strings.NewReader(line)).Decode(u); err != nil {
			return err
		}
	case SpecRevision:
		// v2
		if err = json.NewDecoder(strings.NewReader(strings.TrimPrefix(line, unitSpecPrefix))).Decode(u); err != nil {
			return err
		}
	}
	src, err := ioutil.ReadFile(u.UnitFile.Path)
	if err != nil {
		return err
	}
	u.UnitFile.Source = string(src)
	return nil
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
	return UnitFile{
		SystemPaths: paths,
		Path:        filepath.Join(basePath, unitName),
	}
}

func (f *UnitFile) Read() (err error) {
	src, err := ioutil.ReadFile(f.Path)
	if err != nil {
		return err
	}
	f.Source = string(src)
	return nil
}

func (f *UnitFile) Write() (err error) {
	return ioutil.WriteFile(f.Path, []byte(f.Source), 755)
}

func (f *UnitFile) UnitName() (res string) {
	return filepath.Base(f.Path)
}

func (f *UnitFile) IsRuntime() (res bool) {
	return filepath.Dir(f.Path) == f.SystemPaths.Runtime
}
