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
Search results will not print when auto download enabled.
Auto download calls "qbit torrent add ...", which means it also reads default save values of torrent part on config file
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

	// auto download flags
	searchCmd.Flags().BoolVar(&autoDownload, "auto-download", false, "Attention: if true, it will auto download all the torrents that filter by torrent-regex")
	searchCmd.Flags().BoolVar(&autoMM, "auto-management", true, "whether enable torrent auto management default is true, valid only when auto download enabled")
	// auto download save flags
	searchCmd.Flags().StringVar(&saveCategory, "save-category", "", "torrent save category, valid only when auto download enabled")
	searchCmd.Flags().StringVar(&torrentRegex, "torrent-regex", "", "torrent file name filter")
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

		results := api.SearchDetails(1*time.Second, result.ID, params)
		if results == nil {
			return nil
		}

		if autoDownload {
			download(&results, torrentRegex, autoMM, savePath, saveCategory, saveTags)
		} else {
			fmt.Printf("total search result size: %d\n", len(results))
			for _, r := range results {
				fmt.Printf("{%s}:{%s}\n", r.FileName, r.FileURL)
			}
		}

		return nil
	}

	return searchCmd
}

func download(results *[]api.SearchDetail, regex string, autoMM bool, savePath, saveCategory, saveTags string) {

	var re *regexp.Regexp
	if regex != "" {
		r, err := regexp.Compile(regex)
		if err != nil {
			fmt.Printf("regex: %s compile failed\n", regex)
			return
		}
		re = r
	}

	var urls []string
	for _, r := range *results {
		if re != nil && re.MatchString(r.FileName) {
			urls = append(urls, r.FileURL)
		}
	}

	if len(urls) == 0 {
		fmt.Printf("auto download no results found by regex: %s\n", regex)
		return
	}

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
