package lib

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type StaticBuffers [][]byte

func (s *StaticBuffers) ReadFiles(paths ...string) (err error) {
	var failures []error
	for _, path := range paths {
		if failure := s.read(path); failure != nil {
			failures = append(failures, failure)
		}
	}
	if len(failures) > 0 {
		err = fmt.Errorf("%v", failures)
	}
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
