package command

import (
	"github.com/akaspin/cut"
	"github.com/spf13/cobra"
	"io"
)

var V string

type Soil struct {
	*cut.Environment
}

func (c *Soil) Bind(cc *cobra.Command) {
	cc.Use = "soil"
}

func Run(stderr, stdout io.Writer, stdin io.Reader, args ...string) (err error) {
	env := &cut.Environment{
		Stderr: stderr,
		Stdin:  stdin,
		Stdout: stdout,
	}
	configs := &ConfigOptions{}

	cmd := cut.Attach(
		&Soil{env}, []cut.Binder{env},
		cut.Attach(
			&Agent{env, configs}, []cut.Binder{configs},
		),
	)
	cmd.SetArgs(args)
	cmd.SetOutput(stderr)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	err = cmd.Execute()
	return
}