package cmd

import "github.com/spf13/cobra"

func JackettCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "jackett [command]",
		Short: "Manage jackett",
	}

	cmd.AddCommand(JackettFeed())

	return cmd
}
