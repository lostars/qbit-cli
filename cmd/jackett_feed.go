package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"qbit-cli/internal/config"
)

func JackettFeed() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "feed <keywords> [flags]",
		Short: "Add jackett feed to qBittorrent",
		Long: `Make sure your qBittorrent is configured properly for Jackett.
It will add several feeds depend on your keywords size.
Feed name is your keyword.
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires at least one keyword")
			}
			return nil
		},
	}

	var (
		indexer, category, rule string
	)

	cmd.Flags().StringVar(&indexer, "indexer", "", "jackett indexer name(id)")
	cmd.Flags().StringVar(&category, "category", "", "jackett indexer category")

	cmd.Flags().StringVar(&rule, "rule", "", "qBittorrent rule name")

	cmd.RunE = func(c *cobra.Command, args []string) error {
		if indexer == "" {
			return errors.New("--indexer flag is required")
		}

		cfg, err := config.GetConfig()
		if err != nil {
			return err
		}
		if cfg.Jackett.Host == "" || cfg.Jackett.ApiKey == "" {
			return errors.New("jackett host or api key is empty")
		}

		for _, arg := range args {
			url := cfg.Jackett.Host + "/api/v2.0/indexers/" + indexer +
				"/results/torznab/api?t=search&cat=" + category + "&q=" + arg +
				"&apikey=" + cfg.Jackett.ApiKey

			subCmd := SubAdd()
			_ = subCmd.Flags().Set("rule", rule)
			_ = subCmd.Flags().Set("path", arg)
			subCmd.SetArgs([]string{url})
			if err := subCmd.Execute(); err != nil {
				fmt.Printf("%s add sub error: %s\n", arg, err)
			}
		}

		return nil
	}

	return cmd
}
