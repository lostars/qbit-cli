package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/url"
	"qbit-cli/internal/api"
	"qbit-cli/pkg/utils"
	"strconv"
)

//https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-torrent-list

func TorrentList() *cobra.Command {
	listCmd := &cobra.Command{
		Use:     "list [flags]",
		Short:   "List torrents",
		Example: `qbit torrent list --filter=downloading --category=abc`,
	}

	var (
		filter, category, hashes, tag string
		limit, offset                 uint32
	)

	listCmd.Flags().StringVar(&filter, "filter", "", `state filter:
all, downloading, seeding, completed, stopped, active, inactive, running, 
stalled, stalled_uploading, stalled_downloading, errored`)
	listCmd.Flags().StringVar(&category, "category", "", "category filter")
	listCmd.Flags().StringVar(&tag, "tag", "", "tag filter")
	listCmd.Flags().StringVar(&hashes, "hashes", "", "hash filter separated by |")
	listCmd.Flags().Uint32Var(&limit, "limit", 0, "results limit")
	listCmd.Flags().Uint32Var(&offset, "offset", 0, "results offset")

	listCmd.RunE = func(cmd *cobra.Command, args []string) error {
		var params = url.Values{}
		if filter != "" {
			params.Set("filter", filter)
		}
		if category != "" {
			// category must be encoded
			params.Set("category", url.QueryEscape(category))
		}
		if tag != "" {
			// tag must be encoded
			params.Set("tag", url.QueryEscape(tag))
		}
		if hashes != "" {
			params.Set("hashes", hashes)
		}
		if limit > 0 {
			params.Set("limit", strconv.FormatUint(uint64(limit), 10))
		}
		if offset > 0 {
			params.Set("offset", strconv.FormatUint(uint64(offset), 10))
		}

		torrentList, err := api.TorrentList(params)
		if err != nil {
			return err
		}

		fmt.Printf("total size: %d\n", len(torrentList))

		headers := []string{"name", "hash", "category", "tags", "state", "progress"}
		var data [][]string
		for _, t := range torrentList {
			data = append(data, []string{utils.TruncateStringFromStart(t.Name, 30), t.Hash, t.Category, t.Tags, t.State, strconv.FormatUint(uint64(t.Progress), 10)})
		}
		utils.PrintList(headers, &data)
		return nil
	}

	return listCmd
}
