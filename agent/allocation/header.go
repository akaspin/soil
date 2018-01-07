package allocation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mitchellh/hashstructure"
	"strings"
)

const podSpecPrefix = "### POD"

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

func (h *Header) Unmarshal(src string, paths SystemPaths) (err error) {
	split := strings.Split(src, "\n")
	// extract header
	for _, line := range split {
		if strings.HasPrefix(line, podSpecPrefix) {
			var jsonSrc string
			if _, err = fmt.Sscanf(split[0], "### POD %s %s", &h.Name, &jsonSrc); err != nil {
				return
			}
			if err = json.Unmarshal([]byte(jsonSrc), &h); err != nil {
				return
			}
			return
		}
	}
	return
}

func (h *Header) Marshal(name string, units []*Unit, blobs []*Blob, resources []*Resource, providers ProviderSlice) (res string, err error) {
	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)

	if err = (&SpecMeta{
		Revision: SpecRevision,
	}).Marshal(buf); err != nil {
		return
	}

	if _, err = fmt.Fprintf(buf, "### POD %s ", name); err != nil {
		return
	}
	if err = encoder.Encode(map[string]interface{}{
		"PodMark":   h.PodMark,
		"AgentMark": h.AgentMark,
		"Namespace": h.Namespace,
	}); err != nil {
		return
	}
	for _, u := range units {
		if err = u.MarshalLine(buf); err != nil {
			return
		}
	}

	for _, b := range blobs {
		if err = b.MarshalLine(buf); err != nil {
			return
		}
	}
	for _, resource := range resources {
		if err = resource.MarshalLine(buf); err != nil {
			return
		}
	}
	for _, provider := range providers {
		if err = provider.MarshalLine(buf); err != nil {
			return
		}
	}
	res = buf.String()
	return
}
