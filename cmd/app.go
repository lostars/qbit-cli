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
		info, err := api.QbitAppBuildInfo()
		if err != nil {
			return err
		}
		appVersion, e := api.QbitAppVersion()
		if e != nil {
			return e
		}
		apiVersion, er := api.QbitApiVersion()
		if er != nil {
			return er
		}
		info.AppVersion = appVersion
		info.WebApiVersion = apiVersion
		fmt.Println(info)
		return nil
	}

	return cmd
}
