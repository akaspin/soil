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
	return err.(*multierror.Error).ErrorOrNil()
}

// Get readers for each buffer
func (s StaticBuffers) GetReaders() (res []io.Reader) {
	for _, buf := range s {
		res = append(res, bytes.NewReader(buf))
	}
	return res
}

func (s *StaticBuffers) read(path string) (err error) {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	*s = append(*s, data)
	return nil
}
