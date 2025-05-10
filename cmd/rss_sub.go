package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"qbit-cli/internal/api"
)

func RssSub() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "sub [command]",
		Short: "Manage subscriptions",
		Long:  `If you want add feed, use "rss feed" enhanced command`,
	}

	cmd.AddCommand(SubList())
	cmd.AddCommand(RemoveSub())

	return cmd
}

func SubList() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "Manage subscription",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		results := api.RssAllItems(false)
		if results == nil {
			return nil
		}
		for k, v := range results {
			fmt.Printf("[%s]: [%s]\n", k, v.URL)
		}
		return nil
	}

	return cmd
}

func RemoveSub() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "rm <rss>",
		Short: "remove feeds by name",
		Args: func(c *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("rm requires at least one rss name")
			}
			return nil
		},
	}

	cmd.RunE = func(c *cobra.Command, args []string) error {
		results := api.RssAllItems(false)
		if results == nil {
			return errors.New("no feeds found to remove")
		}

		for _, url := range args {
			if url == "" {
				continue
			}
			_, exists := results[url]
			if !exists {
				fmt.Printf("[%s] not exists, remove failed\n", url)
				continue
			}
			err := api.RssRmSub(url)
			if err != nil {
				fmt.Printf("[%s] remov failed: %v\n", url, err)
				continue
			}
		}
		return nil
	}

	return cmd
}
