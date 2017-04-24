package allocation

import (
	"encoding/json"
	"fmt"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/soil/manifest"
	"github.com/mitchellh/hashstructure"
	"strings"
)

const podUnitTemplate = `
[Unit]
Description=${pod.name}
Before=${pod.units}
[Service]
${agent.pod.exec}
[Install]
WantedBy=${pod.target}
`

type Pod struct {
	*PodHeader
	*File
	Units []*Unit
}

func NewFromManifest(namespace string, m *manifest.Pod, env *agent.Environment) (p *Pod, err error) {
	p = &Pod{
		PodHeader: &PodHeader{
			Name:      m.Name,
			PodMark:   m.Mark(),
			AgentMark: env.Mark(),
			Namespace: namespace,
		},
		File: NewFile(fmt.Sprintf("pod-%s-%s.service", namespace, m.Name), m.Runtime),
	}
	var names []string
	for _, u := range m.Units {
		pu := &Unit{
			UnitHeader: &UnitHeader{
				Permanent:  u.Permanent,
				Transition: u.Transition,
			},
			File: NewFile(u.Name, m.Runtime),
		}
		pu.Source = env.Interpolate(u.Source)
		p.Units = append(p.Units, pu)
		names = append(names, u.Name)
	}
	p.Source, err = p.PodHeader.Marshal(p.Name, p.Units)
	p.Source += env.Interpolate(podUnitTemplate, map[string]string{
			"pod.units":    strings.Join(names, " "),
			"pod.name":   m.Name,
			"pod.target": m.Target,
		})

	return
}

func (p *Pod) PodUnit() (res *Unit) {
	res = &Unit{
		File: p.File,
		UnitHeader: &UnitHeader{
			Permanent: true,
			Transition: manifest.Transition{
				Create:  "start",
				Update:  "restart",
				Destroy: "stop",
			},
		},
	}
	return
}

type PodHeader struct {
	Name      string
	PodMark   uint64
	AgentMark uint64
	Namespace string
}


func (h *PodHeader) Mark() (res uint64) {
	res, _ = hashstructure.Hash(h, nil)
	return
}

func (h *PodHeader) Unmarshal(src string) (units []*Unit, err error) {
	split := strings.Split(src, "\n")
	// extract header
	var jsonSrc string
	if _, err = fmt.Sscanf(split[0], "### POD %s %s", &h.Name, &jsonSrc); err != nil {
		return
	}
	if err = json.Unmarshal([]byte(jsonSrc), &h); err != nil {
		return
	}
	for _, line := range split[1:] {
		if !strings.HasPrefix(line, "### UNIT") {
			break
		}
		u := &Unit{
			File:       &File{},
			UnitHeader: &UnitHeader{},
		}
		if _, err = fmt.Sscanf(line, "### UNIT %s %s", &u.File.Path, &jsonSrc); err != nil {
			return
		}
		if err = json.Unmarshal([]byte(jsonSrc), &u); err != nil {
			return
		}
		units = append(units, u)
	}
	return
}

func (h *PodHeader) Marshal(name string, units []*Unit) (res string, err error) {
	var jsonRes []byte

	if jsonRes, err = json.Marshal(map[string]interface{}{
		"PodMark":   h.PodMark,
		"AgentMark": h.AgentMark,
		"Namespace": h.Namespace,
	}); err != nil {
		return
	}
	res += fmt.Sprintf("### POD %s %s\n", name, string(jsonRes))
	for _, u := range units {
		if jsonRes, err = json.Marshal(&u.UnitHeader); err != nil {
			return
		}
		res += fmt.Sprintf("### UNIT %s %s\n", u.File.Path, string(jsonRes))
	}

	return
}

type Unit struct {
	*File
	*UnitHeader
}

type UnitHeader struct {
	manifest.Transition `json:",squash"`
	Permanent           bool
}
