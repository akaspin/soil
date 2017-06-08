package command

import (
	"fmt"
	"github.com/akaspin/cut"
	"github.com/spf13/cobra"
)

var V string

type Version struct {
	*cut.Environment
}

func (c *Version) Bind(cc *cobra.Command) {
	cc.Use = `version`
	cc.Short = "Print version and exit"
}

func (c *Version) Run(args ...string) (err error) {
	fmt.Fprint(c.Stderr, V)
	return
}
