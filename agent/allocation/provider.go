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

type Providers []*Provider

func (p *Providers) FromManifest(pod manifest.Pod, env map[string]string) (err error) {
	for _, decl := range pod.Providers {
		// clone provider
		v, _ := copystructure.Copy(decl)
		provider := Provider(v.(manifest.Provider))
		*p = append(*p, &provider)
	}
	return
}

func (p *Providers) Append(v ItemUnmarshaller) {
	*p = append(*p, v.(*Provider))
}

type Provider manifest.Provider

func (p *Provider) GetID(parent ...string) string {
	return strings.Join(append(parent, p.Name), ".")
}

// Restore state from header line
func (p *Provider) UnmarshalLine(line string) (err error) {
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
