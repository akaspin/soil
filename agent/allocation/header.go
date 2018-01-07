package allocation

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/hashstructure"
	"io"
	"strings"
)

const podSpecPrefix = "### POD "

type Header struct {
	Name      string
	PodMark   uint64
	AgentMark uint64
	Namespace string
}

func (h *Header) Mark() (res uint64) {
	res, _ = hashstructure.Hash(h, nil)
	return
}

func (h *Header) MarshalLine(w io.Writer) (err error) {
	if _, err = w.Write([]byte(podSpecPrefix)); err != nil {
		return
	}
	err = json.NewEncoder(w).Encode(h)
	return
}

func (h *Header) UnmarshalItem(src string, spec SpecMeta, paths SystemPaths) (err error) {
	for _, line := range strings.Split(src, "\n") {
		if strings.HasPrefix(line, podSpecPrefix) {
			switch spec.Revision {
			case SpecRevision:
				if err = json.NewDecoder(strings.NewReader(strings.TrimPrefix(line, podSpecPrefix))).Decode(h); err != nil {
					return
				}
			default:
				var jsonSrc string
				if _, err = fmt.Sscanf(line, "### POD %s %s", &h.Name, &jsonSrc); err != nil {
					return
				}
				if err = json.Unmarshal([]byte(jsonSrc), &h); err != nil {
					return
				}
			}
			return
		}
	}
	return
}
