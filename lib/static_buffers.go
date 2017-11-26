package lib

import (
	"bytes"
	"github.com/hashicorp/go-multierror"
	"io"
	"io/ioutil"
	"os"
)

type StaticBuffers [][]byte

func (s *StaticBuffers) ReadFiles(paths ...string) (err error) {
	err = &multierror.Error{}
	for _, path := range paths {
		if failure := s.read(path); failure != nil {
			err = multierror.Append(err, failure)
		}
	}
	err = err.(*multierror.Error).ErrorOrNil()
	return
}

// Get readers for each buffer
func (s StaticBuffers) GetReaders() (res []io.Reader) {
	for _, buf := range s {
		res = append(res, bytes.NewReader(buf))
	}
	return
}

func (s *StaticBuffers) read(path string) (err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}
	*s = append(*s, data)
	return
}
