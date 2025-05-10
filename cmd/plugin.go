package cmd

import (
	"github.com/spf13/cobra"
)

func PluginCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "plugin [command]",
		Short: "Manage search plugins",
	}

	cmd.AddCommand(PluginList())

	return cmd
}
