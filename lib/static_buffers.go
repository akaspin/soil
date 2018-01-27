package lib

import (
	"bytes"
	"github.com/akaspin/errslice"
	"io"
	"io/ioutil"
	"os"
)

type StaticBuffers [][]byte

func (s *StaticBuffers) ReadFiles(paths ...string) (err error) {
	for _, path := range paths {
		if failure := s.read(path); failure != nil {
			err = errslice.Append(err, failure)
		}
	}
	return err
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
