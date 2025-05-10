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
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires at least one keyword")
			}
			return nil
		},
	}

	var (
		indexer, category, rule, feedName string
	)

	cmd.Flags().StringVar(&indexer, "indexer", "", "jackett indexer name(id)")
	cmd.Flags().StringVar(&category, "category", "", "jackett indexer category")

	cmd.Flags().StringVar(&rule, "rule", "", "qBittorrent rule name")
	cmd.Flags().StringVar(&feedName, "path", "", "qBittorrent feed name")

	cmd.RunE = func(c *cobra.Command, args []string) error {
		if indexer == "" {
			return errors.New("--indexer flag is required")
		}

		cfg, err := config.GetConfig()
		if err != nil || !cfg.ValidateJackettConfig() {
			return errors.New("get jackett config failed")
		}

		var urls []string
		for _, arg := range args {
			url := cfg.Jackett.Host + "/api/v2.0/indexers/" + indexer +
				"/results/torznab/api?t=search&cat=" + category + "&q=" + arg +
				"&apikey=" + cfg.Jackett.ApiKey
			urls = append(urls, url)
		}

		feed(urls, rule, feedName)

		return nil
	}

	return cmd
}

func feed(urls []string, rule, feedName string) {

	feedCmd := RssFeed()
	_ = feedCmd.Flags().Set("rule", rule)
	_ = feedCmd.Flags().Set("path", feedName)
	feedCmd.SetArgs(urls)

	if err := feedCmd.Execute(); err != nil {
		fmt.Printf("jackett feed error: %v\n", err)
	}
}
