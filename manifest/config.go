package manifest

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type ConfigReader struct {
	chunks [][]byte
}

func NewConfigReader(paths ...string) (c *ConfigReader, err error) {
	c = &ConfigReader{}
	var failures []error
	for _, path := range paths {
		if failure := c.read(path); failure != nil {
			failures = append(failures, failure)
		}
	}
	if len(failures) > 0 {
		err = fmt.Errorf("%v", failures)
	}
	return
}

func (c *ConfigReader) GetReaders() (res []io.Reader) {
	for _, chunk := range c.chunks {
		res = append(res, bytes.NewReader(chunk))
	}
	return
}

func (c *ConfigReader) read(path string) (err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}
	c.chunks = append(c.chunks, data)
	return
}
