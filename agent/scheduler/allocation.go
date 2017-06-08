package scheduler

import (
	"encoding/json"
	"fmt"
	"github.com/akaspin/soil/manifest"
	"github.com/mitchellh/hashstructure"
	"io/ioutil"
	"os"
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
	systemdLocalDir   = "/etc/systemd/system"
	systemdRuntimeDir = "/run/systemd/system"
)

// Allocation represents pod allocated on agent
type Allocation struct {
	*AllocationHeader
	*AllocationFile
	Units []*AllocationUnit
	Blobs []*AllocationBlob
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
	fileHashes := map[string]string{}
	for _, b := range m.Blobs {
		ab := &AllocationBlob{
			Name:        b.Name,
			Permissions: b.Permissions,
			Leave:       b.Leave,
			Source:      manifest.Interpolate(b.Source, env),
		}
		p.Blobs = append(p.Blobs, ab)
		fileHash, _ := hashstructure.Hash(ab.Source, nil)
		fileHashes[fmt.Sprintf("blob.%s", strings.Replace(strings.Trim(ab.Name, "/"), "/", "-", -1))] = fmt.Sprintf("%d", fileHash)
	}
	var unitNames []string
	for _, u := range m.Units {
		pu := &AllocationUnit{
			AllocationUnitHeader: &AllocationUnitHeader{
				Permanent:  u.Permanent,
				Transition: u.Transition,
			},
			AllocationFile: NewFile(u.Name, m.Runtime),
		}
		pu.Source = manifest.Interpolate(u.Source, fileHashes, env)
		p.Units = append(p.Units, pu)
		unitNames = append(unitNames, u.Name)
	}

	p.Source, err = p.AllocationHeader.Marshal(p.Name, p.Units, p.Blobs)
	p.Source += manifest.Interpolate(podUnitTemplate, map[string]string{
		"pod.units":  strings.Join(unitNames, " "),
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
	if res.Units, res.Blobs, err = res.AllocationHeader.Unmarshal(res.AllocationFile.Source); err != nil {
		return
	}

	for _, u := range res.Units {
		if err = u.AllocationFile.Read(); err != nil {
			return
		}
	}
	for _, b := range res.Blobs {
		if err = b.Read(); err != nil {
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

func (h *AllocationHeader) Unmarshal(src string) (units []*AllocationUnit, blobs []*AllocationBlob, err error) {
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
		if strings.HasPrefix(line, "### UNIT") {
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
		if strings.HasPrefix(line, "### BLOB") {
			b := &AllocationBlob{}
			if _, err = fmt.Sscanf(line, "### BLOB %s %s", &b.Name, &jsonSrc); err != nil {
				return
			}
			if err = json.Unmarshal([]byte(jsonSrc), &b); err != nil {
				return
			}
			blobs = append(blobs, b)
		}
	}
	return
}

func (h *AllocationHeader) Marshal(name string, units []*AllocationUnit, blobs []*AllocationBlob) (res string, err error) {
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
	for _, b := range blobs {
		if jsonRes, err = json.Marshal(map[string]interface{}{
			"Permissions": b.Permissions,
			"Leave":       b.Leave,
		}); err != nil {
			return
		}
		res += fmt.Sprintf("### BLOB %s %s\n", b.Name, string(jsonRes))
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

type AllocationBlob struct {
	Name        string
	Permissions int
	Leave       bool
	Source      string
}

func (b *AllocationBlob) Read() (err error) {
	src, err := ioutil.ReadFile(b.Name)
	if err != nil {
		return
	}
	b.Source = string(src)
	return
}

func (b *AllocationBlob) Write() (err error) {
	if err = os.MkdirAll(filepath.Dir(b.Name), os.FileMode(b.Permissions)); err != nil {
		return
	}
	err = ioutil.WriteFile(b.Name, []byte(b.Source), os.FileMode(b.Permissions))
	return
}
