package allocation

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Blob struct {
	Name        string
	Permissions int
	Leave       bool
	Source      string
}

func (b *Blob) Read() (err error) {
	src, err := ioutil.ReadFile(b.Name)
	if err != nil {
		return
	}
	b.Source = string(src)
	return
}

func (b *Blob) Write() (err error) {
	if err = os.MkdirAll(filepath.Dir(b.Name), os.FileMode(b.Permissions)); err != nil {
		return
	}
	err = ioutil.WriteFile(b.Name, []byte(b.Source), os.FileMode(b.Permissions))
	return
}

func (b *Blob) MarshalHeader(w io.Writer, encoder *json.Encoder) (err error) {
	if _, err = fmt.Fprintf(w, "### BLOB %s ", b.Name); err != nil {
		return
	}
	err = encoder.Encode(map[string]interface{}{
		"Permissions": b.Permissions,
		"Leave":       b.Leave,
	})
	return
}
