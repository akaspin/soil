package cut

import (
	"github.com/spf13/cobra"
	"os"
	"io"
)

// streams

/* Environment Binder may be used to automatic bound to OS streams
and working directory.

	func run(args []string, stdin io.Reader, stdout, stderr io.Writer) {
		env := &Environment{
			Stdin:  stdin,
			Stdout: stdout,
			Stderr: stderr,
		}
		root := Attach(
			&RootCmd{}, []Binder{env},
				Attach(&MyCommand{Environment: env}, nil),
		)
		root.SetArgs(args)
		err = root.Execute()
	}

*/
type Environment struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	WD     string
}

func (e *Environment) Bind(cc *cobra.Command) {
	cc.SetOutput(e.Stderr)
	e.WD, _ = os.Getwd()
}
