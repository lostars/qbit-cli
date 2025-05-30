package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"qbit-cli/internal/api"
)

func RssSub() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "sub [command]",
		Short: "Manage subscriptions",
	}

	cmd.AddCommand(SubList())
	cmd.AddCommand(DeleteSub())
	cmd.AddCommand(SubAdd())

	return cmd
}

func SubAdd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add <url> [flags]",
		Short:   "Add subscription",
		Example: "add url --rule=test --path=movie",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("you must provide a rss url")
			}
			return nil
		},
	}

	var (
		rule, path string
	)

	cmd.Flags().StringVar(&rule, "rule", "", "attached rule name")
	cmd.Flags().StringVar(&path, "path", "", "feed name. name will auto generated by url index")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {

		url := args[0]
		if err := api.RssAddSub(url, path); err != nil {
			return err
		}

		var rssRule *api.RssRule
		if rule != "" {
			ruleMap, err := api.RssRuleList()
			if err != nil {
				return err
			}
			if r := ruleMap[rule]; r != nil {
				r.AffectedFeeds = append(r.AffectedFeeds, url)
				rssRule = r
				if err := api.RssSetRule(rule, rssRule); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("%s add failed by: [%s] not found\n", url, rule)
			}
		}
		return nil
	}

	return cmd
}

func SubList() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "Display subscription in formated json",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		results, err := api.RssAllItems(true)
		if err != nil {
			return err
		}
		str, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(str))
		return nil
	}

	return cmd
}

func DeleteSub() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "delete <rss>",
		Short: "Delete feeds by name",
		Args: func(c *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("rm requires at least one rss name")
			}
			return nil
		},
	}

	cmd.RunE = func(c *cobra.Command, args []string) error {
		results, err := api.RssAllItems(false)
		if err != nil {
			return err
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
