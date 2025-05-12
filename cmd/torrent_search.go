package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"net/url"
	"qbit-cli/internal/api"
	"regexp"
	"strconv"
	"time"
)

// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#search

func TorrentSearch() *cobra.Command {

	var searchCmd = &cobra.Command{
		Use:   "search <keyword> [flags]",
		Short: "Search torrents through qBittorrent plugins",
		Long: `Be attention when you enable auto download,
and ensure that torrent-regex works properly to void unnecessary downloads.
Auto download calls "qbit torrent add ...", which means it also reads default save values of torrent part on config file.
This list will show as k:v caused by long magnet display.
`,
		Example: `qbit torrent search <keyword> --category=movie --plugins=bt4g`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a keyword")
			}
			return nil
		},
	}

	var (
		category, plugins                              string
		autoDownload, autoMM                           bool
		torrentRegex, savePath, saveCategory, saveTags string
	)

	// search flags
	searchCmd.Flags().StringVar(&plugins, "plugins", "enabled", `plugins a|b|c, all and enabled also supported.
make sure you plugin is valid and enabled`)
	searchCmd.Flags().StringVar(&category, "category", "all", "category of plugin(define by plugin)")
	searchCmd.Flags().StringVar(&torrentRegex, "torrent-regex", "", "torrent file name filter")

	// auto download flags
	searchCmd.Flags().BoolVar(&autoDownload, "auto-download", false, "Attention: if true, it will auto download all the torrents that filter by torrent-regex")
	searchCmd.Flags().BoolVar(&autoMM, "auto-management", true, "whether enable torrent auto management default is true, valid only when auto download enabled")
	// auto download save flags
	searchCmd.Flags().StringVar(&saveCategory, "save-category", "", "torrent save category, valid only when auto download enabled")
	searchCmd.Flags().StringVar(&savePath, "save-path", "", "torrent save path, valid only when auto download enabled")
	searchCmd.Flags().StringVar(&saveTags, "save-tags", "", "torrent save tags, valid only when auto download enabled")

	searchCmd.RunE = func(cmd *cobra.Command, args []string) error {

		params := url.Values{
			"pattern": {args[0]},
		}

		if plugins != "" {
			params.Set("plugins", plugins)
		}
		if category != "" {
			params.Set("category", category)
		}

		result, err := api.SearchStart(params)
		if err != nil {
			return err
		}

		results, err := api.SearchDetails(1*time.Second, result.ID)
		if err != nil {
			return nil
		}

		var re *regexp.Regexp
		if torrentRegex != "" {
			r, err := regexp.Compile(torrentRegex)
			if err != nil {
				fmt.Printf("regex: %s compile failed\n", torrentRegex)
			} else {
				re = r
			}
		}

		var urls []string
		var printList = make([]*api.SearchDetail, 0, len(results))
		for _, r := range results {
			if re == nil {
				printList = append(printList, r)
			} else {
				if re.MatchString(r.FileName) {
					printList = append(printList, r)
					urls = append(urls, r.FileURL)
				}
			}
		}

		fmt.Printf("total search result size: %d\n", len(printList))
		for _, r := range printList {
			fmt.Printf("%s : %s\n\n", r.FileName, r.FileURL)
		}

		if autoDownload {
			download(urls, autoMM, savePath, saveCategory, saveTags)
		}

		return nil
	}

	return searchCmd
}

func download(urls []string, autoMM bool, savePath, saveCategory, saveTags string) {

	addCmd := TorrentAdd()
	_ = addCmd.Flags().Set("category", saveCategory)
	_ = addCmd.Flags().Set("tags", saveTags)
	_ = addCmd.Flags().Set("auto-management", strconv.FormatBool(autoMM))
	_ = addCmd.Flags().Set("save-path", savePath)
	addCmd.SetArgs(urls)
	if err := addCmd.Execute(); err != nil {
		fmt.Println("auto download failed:", err)
	} else {
		fmt.Printf("auto download %d torrent(s) success\n", len(urls))
	}
}
