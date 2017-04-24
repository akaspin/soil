package allocation

import (
	"io/ioutil"
	"path/filepath"
	"strings"
)

const (
	localDir   = "/usr/lib/systemd/system"
	runtimeDir = "/run/systemd/system"
)

type File struct {
	Path   string
	Source string
}

func NewFile(unitName string, runtime bool) (f *File) {
	basePath := localDir
	if runtime {
		basePath = runtimeDir
	}
	f = &File{
		Path: filepath.Join(basePath, unitName),
	}
	return
}

func (f *File) Read() (err error) {
	src, err := ioutil.ReadFile(f.Path)
	if err != nil {
		return
	}
	f.Source = string(src)
	return
}

func (f *File) Write() (err error) {
	err = ioutil.WriteFile(f.Path, []byte(f.Source), 755)
	return
}

func (f *File) UnitName() (res string) {
	res = filepath.Base(f.Path)
	return
}

func (f *File) IsRuntime() (res bool) {
	res = !strings.HasPrefix(f.Path, localDir)
	return
}
