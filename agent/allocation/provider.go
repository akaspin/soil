package allocation

import (
	"encoding/json"
	"fmt"
	"github.com/akaspin/soil/manifest"
	"github.com/mitchellh/copystructure"
	"io"
	"strings"
)

const (
	providerHeadPrefix = "### PROVIDER "
)

// Allocation providers
type ProviderSlice []*Provider

func (p *ProviderSlice) FromManifest(pod manifest.Pod, env manifest.FlatMap) (err error) {
	for _, decl := range pod.Providers {
		// clone provider
		v, _ := copystructure.Copy(decl)
		provider := Provider(v.(manifest.Provider))
		*p = append(*p, &provider)
	}
	return
}

func (p *ProviderSlice) AppendItem(v ItemUnmarshaller) {
	*p = append(*p, v.(*Provider))
}

type Provider manifest.Provider

func (p *Provider) GetID(parent ...string) string {
	return strings.Join(append(parent, p.Name), ".")
}

// Restore state from header line
func (p *Provider) UnmarshalItem(line string, paths SystemPaths) (err error) {
	err = json.Unmarshal([]byte(strings.TrimPrefix(line, providerHeadPrefix)), p)
	return
}

func (p *Provider) MarshalLine(w io.Writer) (err error) {
	if _, err = fmt.Fprintf(w, "%s", providerHeadPrefix); err != nil {
		return
	}
	err = json.NewEncoder(w).Encode(p)
	return
}

func (p *Provider) Clone() (res *Provider) {
	r1, _ := copystructure.Copy(p)
	res = r1.(*Provider)
	return
}
