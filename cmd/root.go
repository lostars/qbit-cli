package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"qbit-cli/internal/config"
	_ "qbit-cli/internal/job"
)

func Execute(version string) {
	var (
		showVersion, debugMode bool
	)

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
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if debugMode {
				config.Debug = true
				log.SetFlags(log.LstdFlags | log.Lshortfile)
			} else {
				log.SetFlags(0)
				log.SetOutput(io.Discard)
			}
			return
		},
	}

	rootCmd.PersistentFlags().StringVarP(&config.CfgPath, "config", "c", "", "qbit config file path")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "qbit cli version")
	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "d", false, "enable debug")

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
