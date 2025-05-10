package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"qbit-cli/internal/config"
)

func Execute() {

	rootCmd := &cobra.Command{
		Use:   "qbit",
		Short: "qbit is a CLI for qbittorrent",
		Long: `Developed and tested on qBittorrent webui api 5.0.
You may found working properly on other version of qBittorrent.
You can find webui api here:
https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)
`,
		SilenceUsage: true,
	}

	rootCmd.PersistentFlags().StringVarP(&config.CfgPath, "config", "c", "",
		"qbit config file path(default is config.yaml at the same directory as executable file)")

	rootCmd.AddCommand(SearchCmd())
	rootCmd.AddCommand(TorrentCmd())
	rootCmd.AddCommand(RenameCmd())
	rootCmd.AddCommand(RssCmd())
	rootCmd.AddCommand(PluginCmd())
	rootCmd.AddCommand(JackettCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
