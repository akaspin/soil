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
	blobSpecPrefix = "### BLOB "
	blobV2Prefix   = "### BLOB_V2 "
)

type BlobSlice []*Blob

func (s *BlobSlice) GetEmpty(paths SystemPaths) (empty Asset) {
	return &Blob{
		Permissions: 0644,
	}
}

func (s *BlobSlice) GetVersionPrefix(v string) (p string) {
	return blobSpecPrefix
}

func (s *BlobSlice) AppendItem(v Asset) {
	*s = append(*s, v.(*Blob))
}

type Blob struct {
	Name        string
	Permissions int    `json:",omitempty"`
	Leave       bool   `json:",omitempty"`
	Source      string `json:"-"`
}

func (b *Blob) MarshalSpec(w io.Writer) (err error) {
	if _, err = w.Write([]byte(blobSpecPrefix)); err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(b)
}

// Unmarshal blob item from manifest. Line may be in two revisions:
func (b *Blob) UnmarshalSpec(line string, spec Spec, paths SystemPaths) (err error) {
	switch spec.Revision {
	case "":
		if _, err = fmt.Sscanf(line, "### BLOB %s", &b.Name); err != nil {
			return err
		}
		line = strings.TrimPrefix(line, fmt.Sprintf("%s%s ", blobSpecPrefix, b.Name))
		if err = json.NewDecoder(strings.NewReader(line)).Decode(b); err != nil {
			return err
		}
	case SpecRevision:
		// v2
		if err = json.NewDecoder(strings.NewReader(strings.TrimPrefix(line, blobSpecPrefix))).Decode(b); err != nil {
			return err
		}
	}
	src, err := ioutil.ReadFile(b.Name)
	if err != nil {
		return err
	}
	b.Source = string(src)
	return nil
}

func (b *Blob) Write() (err error) {
	if err = os.MkdirAll(filepath.Dir(b.Name), os.FileMode(b.Permissions)); err != nil {
		return err
	}
	return ioutil.WriteFile(b.Name, []byte(b.Source), os.FileMode(b.Permissions))
}
