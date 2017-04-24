package command

import "github.com/spf13/cobra"

type ConfigOptions struct {
	ConfigPath []string
}

func (o *ConfigOptions) Bind(cc *cobra.Command) {
	cc.Flags().StringArrayVarP(&o.ConfigPath, "config", "c", nil, "configuration file")
}

