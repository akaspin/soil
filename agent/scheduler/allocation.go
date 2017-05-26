package scheduler

import (
	"encoding/json"
	"fmt"
	"github.com/akaspin/soil/manifest"
	"github.com/mitchellh/hashstructure"
	"io/ioutil"
	"path/filepath"
	"strings"
)

const (
	podUnitTemplate = `
[Unit]
Description=${pod.name}
Before=${pod.units}
[Service]
${agent.pod_exec}
[Install]
WantedBy=${pod.target}
`
	systemdLocalDir   = "/usr/lib/systemd/system"
	systemdRuntimeDir = "/run/systemd/system"
)

// Allocation represents pod allocated on agent
type Allocation struct {
	*AllocationHeader
	*AllocationFile
	Units []*AllocationUnit
}

func NewAllocationFromManifest(m *manifest.Pod, env map[string]string, mark uint64) (p *Allocation, err error) {
	p = &Allocation{
		AllocationHeader: &AllocationHeader{
			Name:      m.Name,
			PodMark:   m.Mark(),
			AgentMark: mark,
			Namespace: m.Namespace,
		},
		AllocationFile: NewFile(fmt.Sprintf("pod-%s-%s.service", m.Namespace, m.Name), m.Runtime),
	}
	var names []string
	for _, u := range m.Units {
		pu := &AllocationUnit{
			AllocationUnitHeader: &AllocationUnitHeader{
				Permanent:  u.Permanent,
				Transition: u.Transition,
			},
			AllocationFile: NewFile(u.Name, m.Runtime),
		}
		pu.Source = manifest.Interpolate(u.Source, env)
		p.Units = append(p.Units, pu)
		names = append(names, u.Name)
	}
	p.Source, err = p.AllocationHeader.Marshal(p.Name, p.Units)
	p.Source += manifest.Interpolate(podUnitTemplate, map[string]string{
			"pod.units":    strings.Join(names, " "),
			"pod.name":   m.Name,
			"pod.target": m.Target,
		}, env)

	return
}

func NewAllocationFromSystemD(path string) (res *Allocation, err error) {
	res = &Allocation{
		AllocationFile: &AllocationFile{
			Path: path,
		},
		AllocationHeader: &AllocationHeader{},
	}
	if err = res.AllocationFile.Read(); err != nil {
		return
	}
	if res.Units, err = res.AllocationHeader.Unmarshal(res.AllocationFile.Source); err != nil {
		return
	}

	for _, u := range res.Units {
		if err = u.AllocationFile.Read(); err != nil {
			return
		}
	}
	return
}

func (p *Allocation) PodUnit() (res *AllocationUnit) {
	res = &AllocationUnit{
		AllocationFile: p.AllocationFile,
		AllocationUnitHeader: &AllocationUnitHeader{
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

type AllocationHeader struct {
	Name      string
	PodMark   uint64
	AgentMark uint64
	Namespace string
}


func (h *AllocationHeader) Mark() (res uint64) {
	res, _ = hashstructure.Hash(h, nil)
	return
}

func (h *AllocationHeader) Unmarshal(src string) (units []*AllocationUnit, err error) {
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
		u := &AllocationUnit{
			AllocationFile:       &AllocationFile{},
			AllocationUnitHeader: &AllocationUnitHeader{},
		}
		if _, err = fmt.Sscanf(line, "### UNIT %s %s", &u.AllocationFile.Path, &jsonSrc); err != nil {
			return
		}
		if err = json.Unmarshal([]byte(jsonSrc), &u); err != nil {
			return
		}
		units = append(units, u)
	}
	return
}

func (h *AllocationHeader) Marshal(name string, units []*AllocationUnit) (res string, err error) {
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
		if jsonRes, err = json.Marshal(&u.AllocationUnitHeader); err != nil {
			return
		}
		res += fmt.Sprintf("### UNIT %s %s\n", u.AllocationFile.Path, string(jsonRes))
	}

	return
}

type AllocationUnit struct {
	*AllocationFile
	*AllocationUnitHeader
}

type AllocationUnitHeader struct {
	manifest.Transition `json:",squash"`
	Permanent           bool
}

type AllocationFile struct {
	Path   string
	Source string
}

func NewFile(unitName string, runtime bool) (f *AllocationFile) {
	basePath := systemdLocalDir
	if runtime {
		basePath = systemdRuntimeDir
	}
	f = &AllocationFile{
		Path: filepath.Join(basePath, unitName),
	}
	return
}

func (f *AllocationFile) Read() (err error) {
	src, err := ioutil.ReadFile(f.Path)
	if err != nil {
		return
	}
	f.Source = string(src)
	return
}

func (f *AllocationFile) Write() (err error) {
	err = ioutil.WriteFile(f.Path, []byte(f.Source), 755)
	return
}

func (f *AllocationFile) UnitName() (res string) {
	res = filepath.Base(f.Path)
	return
}

func (f *AllocationFile) IsRuntime() (res bool) {
	res = !strings.HasPrefix(f.Path, systemdLocalDir)
	return
}

func AllocationToString(p *Allocation) (res string) {
	if p == nil {
		res = "<nil>"
		return
	}
	res = fmt.Sprintf("%v", p.AllocationHeader)
	return
}
