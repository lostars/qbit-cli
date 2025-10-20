package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"qbit-cli/internal/api"
)

func PluginCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "plugin [command]",
		Short: "Manage search plugins",
	}

	cmd.AddCommand(PluginList())
	cmd.AddCommand(InstallPlugin())
	cmd.AddCommand(UninstallPlugin())
	cmd.AddCommand(EnablePlugin())
	cmd.AddCommand(UpdatePlugin())

	return cmd
}

func InstallPlugin() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "install <url>...",
		Short: "Install plugins",
		Args: func(c *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("plugin install requires at least one url")
			}
			return nil
		},
	}
	cmd.RunE = func(c *cobra.Command, args []string) error {
		err := api.InstallPlugin(args)
		if err != nil {
			return err
		}
		return nil
	}

	return cmd
}

func UninstallPlugin() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "uninstall <name>...",
		Short: "Uninstall plugins. API seems not working...",
		Args: func(c *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("plugin uninstall requires at least one name")
			}
			return nil
		},
	}
	cmd.RunE = func(c *cobra.Command, args []string) error {
		err := api.UninstallPlugin(args)
		if err != nil {
			return err
		}
		return nil
	}
	return cmd
}

func EnablePlugin() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "enable <name>...",
		Short: "Manage plugin status",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires at least one name")
			}
			return nil
		},
	}

	var enable bool
	cmd.Flags().BoolVar(&enable, "enable", true, "enable plugin")

	cmd.RunE = func(c *cobra.Command, args []string) error {
		err := api.EnablePlugin(args, enable)
		if err != nil {
			return err
		}
		return nil
	}
	return cmd
}

func UpdatePlugin() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "update",
		Short: "Update all plugins",
	}
	cmd.RunE = func(c *cobra.Command, args []string) error {
		if err := api.UpdatePlugin(); err != nil {
			return err
		}
		return nil
	}
	return cmd
}
