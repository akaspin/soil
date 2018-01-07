package allocation

import (
	"encoding/json"
	"io"
	"strings"
)

const (
	specRevisionPrefix = "### SOIL "
	SpecRevision       = "1.0"
)

type SpecMeta struct {
	Revision string
}

func (s *SpecMeta) Unmarshal(src string) (err error) {
	for _, line := range strings.Split(src, "\n") {
		if strings.HasPrefix(line, specRevisionPrefix) {
			err = json.NewDecoder(strings.NewReader(strings.TrimSpace(strings.TrimPrefix(line, specRevisionPrefix)))).Decode(&s)
			return
		}
	}
	return
}

func (s *SpecMeta) Marshal(w io.Writer) (err error) {
	if _, err = w.Write([]byte(specRevisionPrefix)); err != nil {
		return
	}
	err = json.NewEncoder(w).Encode(s)
	return
}
