package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"qbit-cli/internal/api"
	"strings"
)

func RssRule() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "rule [command]",
		Short: "Manage RSS rules",
	}

	cmd.AddCommand(RuleList())

	return cmd
}

func RuleList() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "RSS rule list display as formated json",
	}

	var filter string
	cmd.Flags().StringVar(&filter, "filter", "", "filter rule name")

	cmd.RunE = func(c *cobra.Command, args []string) error {

		ruleMap, err := api.RssRuleList()
		if err != nil {
			return err
		}
		if filter != "" {
			for name, rule := range ruleMap {
				if strings.Contains(name, filter) {
					data, _ := json.MarshalIndent(rule, "", "  ")
					fmt.Println(string(data))
				}
			}
			return nil
		}
		data, err := json.MarshalIndent(ruleMap, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))

		return nil
	}

	return cmd
}
