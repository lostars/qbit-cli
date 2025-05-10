package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"qbit-cli/internal/api"
	"regexp"
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
		plugins := api.SearchPlugins()
		if plugins == nil {
			return nil
		}

		var re *regexp.Regexp
		if filter != "" {
			re = regexp.MustCompile(filter)
		}
		for _, plugin := range *plugins {
			if enabled && !plugin.Enabled {
				continue
			}
			if re != nil {
				if re.MatchString(plugin.Name) {
					printPlugin(plugin)
				}
			} else {
				if filter != "" {
					contains := strings.Contains(plugin.Name, filter) || strings.Contains(plugin.FullName, filter)
					if contains {
						printPlugin(plugin)
					}
				} else {
					printPlugin(plugin)
				}
			}
		}

		return nil
	}

	return cmd
}

func printPlugin(plugin api.SearchPlugin) {
	fmt.Printf("name:[%s], fullName:[%s], enabled:[%v], url:[%s]\n",
		plugin.Name, plugin.FullName, plugin.Enabled, plugin.Url)
}
