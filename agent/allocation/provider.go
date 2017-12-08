package allocation

import (
	"encoding/json"
	"fmt"
	"github.com/akaspin/soil/manifest"
	"io"
	"strings"
)

const (
	providerHeadPrefix = "### PROVIDER "
)

type Providers []*Provider

func (p *Providers) Append(v Recoverable) {
	*p = append(*p, v.(*Provider))
}

type Provider manifest.Provider

// Restore state from manifest
func (p *Provider) RestoreState(line string) (err error) {
	err = json.Unmarshal([]byte(strings.TrimPrefix(line, providerHeadPrefix)), p)
	return
}

func (p *Provider) StoreState(w io.Writer) (err error) {
	if _, err = fmt.Fprintf(w, "%s", providerHeadPrefix); err != nil {
		return
	}
	err = json.NewEncoder(w).Encode(p)
	return
}
