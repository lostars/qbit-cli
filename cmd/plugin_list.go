package cmd

import (
	"fmt"
	"github.com/olekukonko/tablewriter/tw"
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
		headers := []string{"name", "enabled", "url", "category"}
		var data = make([][]string, len(printPlugins))
		for i, plugin := range printPlugins {
			cat := make([]string, 0, len(plugin.SupportedCategories))
			for _, v := range plugin.SupportedCategories {
				cat = append(cat, v.ID)
			}
			data[i] = []string{plugin.Name, strconv.FormatBool(plugin.Enabled), plugin.Url, fmt.Sprintf("%s", cat)}
		}
		cell := tw.CellConfig{
			Formatting: tw.CellFormatting{
				MaxWidth:  100,
				AutoWrap:  tw.WrapTruncate,
				Alignment: tw.AlignNone,
			},
		}
		utils.PrintListWithCellConfig(headers, &data, cell)

		return nil
	}

	return cmd
}
