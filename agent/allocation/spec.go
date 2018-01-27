package allocation

import (
	"encoding/json"
	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/copystructure"
	"io"
	"strings"
)

const (
	specRevisionPrefix = "### SOIL "
	SpecRevision       = "1.0"
)

type Spec struct {
	Revision string
}

func (s *Spec) Unmarshal(src string) (err error) {
	for _, line := range strings.Split(src, "\n") {
		if strings.HasPrefix(line, specRevisionPrefix) {
			return json.NewDecoder(strings.NewReader(strings.TrimSpace(strings.TrimPrefix(line, specRevisionPrefix)))).Decode(&s)
		}
	}
	return nil
}

func (s *Spec) Marshal(w io.Writer) (err error) {
	if _, err = w.Write([]byte(specRevisionPrefix)); err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(s)
}

func (s Spec) UnmarshalAssetSlice(paths SystemPaths, v AssetSlice, source string) (err error) {
	err = &multierror.Error{}
	prefix := v.GetVersionPrefix(s.Revision)
	for _, line := range strings.Split(source, "\n") {
		if strings.HasPrefix(line, prefix) {
			cp, _ := copystructure.Copy(v.GetEmpty(paths))
			v1 := cp.(Asset)
			if rErr := v1.UnmarshalSpec(line, s, paths); rErr != nil {
				err = multierror.Append(err, rErr)
				continue
			}
			v.AppendItem(v1)
		}
	}
	return err.(*multierror.Error).ErrorOrNil()
}
