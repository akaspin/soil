package allocation

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	blobV1Prefix = "### BLOB "
	blobV2Prefix = "### BLOB_V2 "
)

type BlobSlice []*Blob

func (s *BlobSlice) AppendItem(v ItemUnmarshaller) {
	*s = append(*s, v.(*Blob))
}

type Blob struct {
	Name        string
	Permissions int    `json:",omitempty"`
	Leave       bool   `json:",omitempty"`
	Source      string `json:"-"`
}

func (b *Blob) MarshalLine(w io.Writer) (err error) {
	if _, err = w.Write([]byte(blobV2Prefix)); err != nil {
		return
	}
	err = json.NewEncoder(w).Encode(b)
	return
}

// Unmarshal blob item from manifest. Line may be in two revisions:
//
//	  v1: ### BLOB <name> <json-spec>
//	  v2: ### BLOB.v2 <json-spec>
func (b *Blob) UnmarshalItem(line string) (err error) {
	switch {
	case strings.HasPrefix(line, blobV1Prefix):
		// v1
		if _, err = fmt.Sscanf(line, "### BLOB %s", &b.Name); err != nil {
			return
		}
		line = strings.TrimPrefix(line, fmt.Sprintf("%s%s ", blobV1Prefix, b.Name))
		if err = json.NewDecoder(strings.NewReader(line)).Decode(b); err != nil {
			return
		}
	case strings.HasPrefix(line, blobV2Prefix):
		// v2
		if err = json.NewDecoder(strings.NewReader(strings.TrimPrefix(line, blobV2Prefix))).Decode(b); err != nil {
			return
		}
	}
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
