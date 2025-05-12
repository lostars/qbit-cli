package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"qbit-cli/internal/config"
)

func Execute(version string) {

	rootCmd := &cobra.Command{
		Use:   "qbit",
		Short: "qbit is a CLI for qBittorrent",
		Long: `Developed and tested on qBittorrent webui api 5.0.

You may found working properly on other version of qBittorrent.
You can find webui api here:
https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)

Default config location:
~/.config/qbit-cli/config.yaml
same as executable file named config.yaml
`,
		SilenceUsage: true,
	}

	var (
		showVersion bool
	)

	rootCmd.PersistentFlags().StringVarP(&config.CfgPath, "config", "c", "", "qbit config file path")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "qbit cli version")

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if showVersion {
			fmt.Println(version)
		}
		return nil
	}

	rootCmd.AddCommand(TorrentCmd())
	rootCmd.AddCommand(RenameCmd())
	rootCmd.AddCommand(RssCmd())
	rootCmd.AddCommand(PluginCmd())
	rootCmd.AddCommand(JackettCmd())
	rootCmd.AddCommand(EmbyCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
