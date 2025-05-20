package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"qbit-cli/internal/api"
)

func AppCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "app",
		Short: "Manage app",
	}

	cmd.AddCommand(AppInfo())

	return cmd
}

func AppInfo() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "info",
		Short: "Show app info",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		fmt.Println(api.GetQbitServerInfo())
		return nil
	}

	return cmd
}
