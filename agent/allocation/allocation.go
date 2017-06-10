package allocation

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

// Pod represents pod allocated on agent
type Pod struct {
	*Header
	*UnitFile
	Units []*Unit
	Blobs []*Blob
}

func NewFromManifest(m *manifest.Pod, env map[string]string, mark uint64) (p *Pod, err error) {
	p = &Pod{
		Header: &Header{
			Name:      m.Name,
			PodMark:   m.Mark(),
			AgentMark: mark,
			Namespace: m.Namespace,
		},
		UnitFile: NewFile(fmt.Sprintf("pod-%s-%s.service", m.Namespace, m.Name), m.Runtime),
	}
	fileHashes := map[string]string{}
	for _, b := range m.Blobs {
		ab := &Blob{
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
		pu := &Unit{
			Transition: &u.Transition,
			UnitFile:   NewFile(u.Name, m.Runtime),
		}
		pu.Source = manifest.Interpolate(u.Source, fileHashes, env)
		p.Units = append(p.Units, pu)
		unitNames = append(unitNames, u.Name)
	}

	p.Source, err = p.Header.Marshal(p.Name, p.Units, p.Blobs)
	p.Source += manifest.Interpolate(podUnitTemplate, map[string]string{
		"pod.units":  strings.Join(unitNames, " "),
		"pod.name":   m.Name,
		"pod.target": m.Target,
	}, env)

	return
}

func NewFromSystemD(path string) (res *Pod, err error) {
	res = &Pod{
		UnitFile: &UnitFile{
			Path: path,
		},
		Header: &Header{},
	}
	if err = res.UnitFile.Read(); err != nil {
		return
	}
	if res.Units, res.Blobs, err = res.Header.Unmarshal(res.UnitFile.Source); err != nil {
		return
	}

	for _, u := range res.Units {
		if err = u.UnitFile.Read(); err != nil {
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

func (p *Pod) GetPodUnit() (res *Unit) {
	res = &Unit{
		UnitFile: p.UnitFile,
		Transition: &manifest.Transition{
			Create:    "start",
			Update:    "restart",
			Destroy:   "stop",
			Permanent: true,
		},
	}
	return
}

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

func (h *Header) Unmarshal(src string) (units []*Unit, blobs []*Blob, err error) {
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
			u := &Unit{
				UnitFile:   &UnitFile{},
				Transition: &manifest.Transition{},
			}
			if _, err = fmt.Sscanf(line, "### UNIT %s %s", &u.UnitFile.Path, &jsonSrc); err != nil {
				return
			}
			if err = json.Unmarshal([]byte(jsonSrc), &u); err != nil {
				return
			}
			units = append(units, u)
		}
		if strings.HasPrefix(line, "### BLOB") {
			b := &Blob{}
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

func (h *Header) Marshal(name string, units []*Unit, blobs []*Blob) (res string, err error) {
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
		if jsonRes, err = json.Marshal(&u.Transition); err != nil {
			return
		}
		res += fmt.Sprintf("### UNIT %s %s\n", u.UnitFile.Path, string(jsonRes))
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

type Unit struct {
	*UnitFile
	*manifest.Transition `json:",squash"`
}

type UnitFile struct {
	Path   string
	Source string
}

func NewFile(unitName string, runtime bool) (f *UnitFile) {
	basePath := systemdLocalDir
	if runtime {
		basePath = systemdRuntimeDir
	}
	f = &UnitFile{
		Path: filepath.Join(basePath, unitName),
	}
	return
}

func (f *UnitFile) Read() (err error) {
	src, err := ioutil.ReadFile(f.Path)
	if err != nil {
		return
	}
	f.Source = string(src)
	return
}

func (f *UnitFile) Write() (err error) {
	err = ioutil.WriteFile(f.Path, []byte(f.Source), 755)
	return
}

func (f *UnitFile) UnitName() (res string) {
	res = filepath.Base(f.Path)
	return
}

func (f *UnitFile) IsRuntime() (res bool) {
	res = !strings.HasPrefix(f.Path, systemdLocalDir)
	return
}

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
