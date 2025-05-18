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
https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-%28qBittorrent-5.0%29

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

	rootCmd.AddCommand(AppCmd())
	rootCmd.AddCommand(TorrentCmd())
	rootCmd.AddCommand(RssCmd())
	rootCmd.AddCommand(PluginCmd())
	rootCmd.AddCommand(JackettCmd())
	rootCmd.AddCommand(EmbyCmd())
	rootCmd.AddCommand(JobCmd())

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("%s\n", r)
			os.Exit(1)
		}
	}()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
