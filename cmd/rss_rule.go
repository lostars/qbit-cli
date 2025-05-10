package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"qbit-cli/internal/api"
)

func RssRule() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "rule [command]",
		Short: "RSS Rule commands",
	}

	cmd.AddCommand(RuleList())

	return cmd
}

func RuleList() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "rss rule list",
	}

	cmd.RunE = func(c *cobra.Command, args []string) error {

		ruleMap := api.RssRuleList()
		if ruleMap == nil {
			return nil
		}

		for ruleName, rule := range ruleMap {
			fmt.Printf("ruleName: [%s] enabled: [%v] mustContain: [%v] mustNotContain: [%v] useRegex: [%v] affectedFeeds: [%v]\n ",
				ruleName, rule.Enabled, rule.MustContain, rule.MustNotContain, rule.UseRegex, rule.AffectedFeeds)
		}

		return nil
	}

	return cmd
}
