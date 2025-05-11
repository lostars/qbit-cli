package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"qbit-cli/internal/api"
	"qbit-cli/pkg/utils"
	"regexp"
	"strconv"
	"strings"
)

func PluginList() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list [flags]",
		Short: "List all plugins",
	}

	var (
		filter  string
		enabled bool
	)

	cmd.Flags().StringVar(&filter, "filter", "", "Filter plugins by name, regex supported")
	cmd.Flags().BoolVar(&enabled, "enabled", false, "list enabled plugins")

	cmd.RunE = func(c *cobra.Command, args []string) error {
		plugins, err := api.SearchPlugins()
		if err != nil {
			return err
		}

		var re *regexp.Regexp
		if filter != "" {
			re = regexp.MustCompile(filter)
		}
		var printPlugins []api.SearchPlugin
		for _, plugin := range *plugins {
			if enabled && !plugin.Enabled {
				continue
			}
			if re != nil {
				if re.MatchString(plugin.Name) {
					printPlugins = append(printPlugins, plugin)
				}
			} else {
				if filter != "" {
					contains := strings.Contains(plugin.Name, filter) || strings.Contains(plugin.FullName, filter)
					if contains {
						printPlugins = append(printPlugins, plugin)
					}
				} else {
					printPlugins = append(printPlugins, plugin)
				}
			}
		}

		fmt.Printf("total plugin size: %d\n", len(printPlugins))
		headers := []string{"name", "fullName", "enabled", "url"}
		var data [][]string
		for _, plugin := range printPlugins {
			data = append(data, []string{plugin.Name, plugin.FullName, strconv.FormatBool(plugin.Enabled), plugin.Url})
		}
		utils.PrintList(headers, &data)

		return nil
	}

	return cmd
}
