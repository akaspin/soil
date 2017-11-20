package command

import (
	"github.com/akaspin/cut"
	"github.com/spf13/cobra"
	"fmt"
	"net/http"
)

type ClientURLOptions struct {
	URL string
	NodeID string
	Redirect bool
}

type ClientOutputOptions struct {
	
}

type Nodes struct {
	*cut.Environment
	*ClientURLOptions
}

func (c *Nodes) Bind(cc *cobra.Command) {
	cc.Use = `nodes`
	cc.Short = "List nodes in cluster"
}

func (c *Nodes) Run(args ...string) (err error) {
	p := fmt.Sprintf("%s/v1/status/nodes", c.ClientURLOptions.URL)
	_, err = http.NewRequest(http.MethodGet, p, nil)
	if err != nil {
		return
	}

	return
}
