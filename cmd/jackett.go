package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"qbit-cli/internal/api"
	"qbit-cli/pkg/utils"
	"strconv"
	"strings"
)

func JackettCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "jackett [command]",
		Short: "Manage Jackett",
	}

	cmd.AddCommand(JackettFeed())
	cmd.AddCommand(JackettSearch())
	cmd.AddCommand(JackettIndexers())

	return cmd
}

func JackettIndexers() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List Jackett indexers. You must provide your Jackett cookie",
	}

	var enabled, jsonFormat bool
	var filter string
	cmd.Flags().BoolVar(&enabled, "enabled", false, "enable the Jackett")
	cmd.Flags().BoolVar(&jsonFormat, "json", false, "display results in json format")
	cmd.Flags().StringVar(&filter, "filter", "", "filter the indexer by id")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		indexers, err := api.JackettIndexers()
		if err != nil {
			return err
		}
		var results = make([]api.JackettIndexer, 0, len(*indexers))
		for _, indexer := range *indexers {
			if enabled && filter != "" {
				if indexer.Configured && strings.Contains(indexer.ID, filter) {
					results = append(results, indexer)
				}
				continue
			}
			if enabled && indexer.Configured {
				results = append(results, indexer)
				continue
			}
			if filter != "" && strings.Contains(indexer.ID, filter) {
				results = append(results, indexer)
				continue
			}
		}

		if jsonFormat {
			str, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(str))
		} else {
			header := []string{"id", "configured", "LANG", "site"}
			var data = make([][]string, 0, len(results))
			for _, indexer := range results {
				data = append(data, []string{indexer.ID, strconv.FormatBool(indexer.Configured),
					indexer.Language, indexer.SiteLink})
			}
			utils.PrintList(header, &data)
		}

		return nil
	}

	return cmd
}
