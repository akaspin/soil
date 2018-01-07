package allocation

import (
	"bytes"
	"fmt"
	"github.com/akaspin/soil/manifest"
	"github.com/mitchellh/hashstructure"
	"strings"
)

const (
	podUnitTemplate = `
[Unit]
Description=${pod.name}
Before=${pod.units}
[Service]
${system.pod_exec}
[Install]
WantedBy=${pod.target}
`
	dirSystemDLocal   = "/etc/systemd/system"
	dirSystemDRuntime = "/run/systemd/system"
)

// Pod represents pod allocated on agent
type Pod struct {
	Header
	UnitFile
	Units     UnitSlice
	Blobs     BlobSlice
	Resources ResourceSlice
	Providers ProviderSlice
}

func NewPod(systemPaths SystemPaths) (p *Pod) {
	p = &Pod{
		UnitFile: UnitFile{
			SystemPaths: systemPaths,
		},
		Header: Header{},
	}
	return
}

func (p *Pod) FromManifest(m *manifest.Pod, env map[string]string) (err error) {
	agentMark, _ := hashstructure.Hash(env, nil)
	p.Header = Header{
		Name:      m.Name,
		PodMark:   m.Mark(),
		AgentMark: agentMark,
		Namespace: m.Namespace,
	}
	e := manifest.FlatMap{
		"pod.name":      m.Name,
		"pod.namespace": m.Namespace,
		"pod.target":    m.Target,
	}.Merge(env)

	p.UnitFile = NewUnitFile(fmt.Sprintf("pod-%s-%s.service", m.Namespace, m.Name), p.SystemPaths, m.Runtime)
	baseEnv := map[string]string{
		"pod.name":      m.Name,
		"pod.namespace": m.Namespace,
	}
	baseSourceEnv := map[string]string{
		"pod.target": m.Target,
	}

	// Blobs
	fileHashes1 := manifest.FlatMap{}
	for _, b := range m.Blobs {
		ab := &Blob{
			Name:        manifest.Interpolate(b.Name, baseEnv),
			Permissions: b.Permissions,
			Leave:       b.Leave,
			Source:      e.Interpolate(b.Source),
		}
		p.Blobs = append(p.Blobs, ab)
		fileHash, _ := hashstructure.Hash(ab.Source, nil)
		fileHashes1[fmt.Sprintf(
			"blob.%s", strings.Replace(strings.Trim(ab.Name, "/"), "/", "-", -1))] = fmt.Sprintf("%d", fileHash)
	}
	e = e.Merge(fileHashes1)

	// Units
	var unitNames []string
	for _, u := range m.Units {
		unitName := manifest.Interpolate(u.Name, baseEnv)
		pu := &Unit{
			Transition: u.Transition,
			UnitFile:   NewUnitFile(unitName, p.SystemPaths, m.Runtime),
		}
		pu.Source = e.Interpolate(u.Source)
		p.Units = append(p.Units, pu)
		unitNames = append(unitNames, unitName)
	}

	p.Resources.FromManifest(*m, env)
	p.Providers.FromManifest(*m, env)

	// marshal pod unit
	var buf bytes.Buffer
	spec := SpecMeta{
		Revision: SpecRevision,
	}
	if err = spec.Marshal(&buf); err != nil {
		return
	}
	if err = p.Header.MarshalLine(&buf); err != nil {
		return
	}
	for _, a := range p.Units {
		if err = a.MarshalLine(&buf); err != nil {
			return
		}
	}
	for _, a := range p.Blobs {
		if err = a.MarshalLine(&buf); err != nil {
			return
		}
	}
	for _, a := range p.Providers {
		if err = a.MarshalLine(&buf); err != nil {
			return
		}
	}
	for _, a := range p.Resources {
		if err = a.MarshalLine(&buf); err != nil {
			return
		}
	}
	if _, err = buf.WriteString(manifest.Interpolate(
		podUnitTemplate,
		baseEnv,
		baseSourceEnv,
		map[string]string{
			"pod.units": strings.Join(unitNames, " "),
		},
		env)); err != nil {
		return
	}
	p.Source = buf.String()
	return
}

func (p *Pod) FromFilesystem(path string) (err error) {
	p.UnitFile.Path = path
	if err = p.UnitFile.Read(); err != nil {
		return
	}

	// get version
	var meta SpecMeta
	if err = meta.Unmarshal(p.UnitFile.Source); err != nil {
		return
	}

	if err = p.Header.UnmarshalItem(p.UnitFile.Source, meta, p.SystemPaths); err != nil {
		return
	}

	src := p.UnitFile.Source
	err = UnmarshalItemSlice(meta, p.SystemPaths, &p.Units, src)
	err = UnmarshalItemSlice(meta, p.SystemPaths, &p.Blobs, src)
	err = UnmarshalItemSlice(meta, p.SystemPaths, &p.Resources, src)
	err = UnmarshalItemSlice(meta, p.SystemPaths, &p.Providers, src)
	return
}

func (p *Pod) GetPodUnit() (res *Unit) {
	res = &Unit{
		UnitFile: p.UnitFile,
		Transition: manifest.Transition{
			Create:    "start",
			Update:    "restart",
			Destroy:   "stop",
			Permanent: true,
		},
	}
	return
}
