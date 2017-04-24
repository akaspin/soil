package cut

import (
	"github.com/spf13/cobra"
	"reflect"
	"strings"
)

type Command interface {}

// Binder can bound some options to cobra.
type Binder interface {
	Bind(cc *cobra.Command)
}

type Runnable interface {

	// Run command with rest of args
	Run(args ...string) (err error)
}

// Attach command. If command implements Binder it will be also evaluated
func Attach(c Command, binders []Binder, cmds ...*cobra.Command) (cc *cobra.Command) {
	cc = &cobra.Command{}
	for _, binder := range binders {
		binder.Bind(cc)
	}
	if c1, ok := c.(Binder); ok {
		c1.Bind(cc)
	}
	if cc.Use == "" {
		// TypeOf().Name() return empty string on pointer types
		cmd := reflect.TypeOf(c).String()
		chunks := strings.Split(cmd, ".")
		cc.Use = strings.ToLower(strings.TrimPrefix(chunks[len(chunks)-1], "*"))
	}

	if c1, ok := c.(Runnable); ok {
		cc.RunE = func(cc *cobra.Command, args []string) (err error) {
			return c1.Run(args...)
		}

	}
	cc.AddCommand(cmds...)
	return
}
