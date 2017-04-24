package cut

import (
	"github.com/spf13/pflag"
	"os"
	"strings"
)

// Override unchanged flags with values from environment
func OverrideEnv(flagSet *pflag.FlagSet, prefix string, flags ...string) (err error) {
	env := os.Environ()
	LOOP:
	for _, f := range flags {
		pf := flagSet.Lookup(f)
		if pf != nil && !pf.Changed {
			envName := strings.ToUpper(prefix + f + "=")
			for _, e := range env {
				if strings.HasPrefix(e, envName) {
					if err = pf.Value.Set(strings.TrimPrefix(e, envName)); err != nil {
						return
					}
					continue LOOP
				}
			}
		}
	}
	return 
}
